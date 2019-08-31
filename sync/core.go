package sync

import (
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/floutil"

	"github.com/oipwg/oip/datastore"
	"github.com/oipwg/oip/events"
	"github.com/oipwg/oip/flo"
)

var (
	IsInitialSync = true
	recentBlocks  = blockBuffer{}
)

func Setup() {
	// ToDo: refresh_interval
	//  https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-indexing-speed.html#_increase_the_refresh_interval
}

func IndexBlockAtHeight(height int64, lb datastore.BlockData) (datastore.BlockData, error) {
	hash, err := flo.GetBlockHash(height)
	if err != nil {
		return lb, err
	}

	b, err := flo.GetBlockVerboseTx(hash)
	if err != nil {
		return lb, err
	}

	var lbt int64
	if lb.Block == nil {
		lbt = b.Time
	} else {
		lbt = lb.Block.Time
	}

	bd := datastore.BlockData{
		Block:             b,
		SecSinceLastBlock: b.Time - lbt,
		Orphaned:          false,
	}

	datastore.AutoBulk.StoreBlock(bd)

	for i := range bd.Block.RawTx {
		rawTx := &bd.Block.RawTx[i]

		feeSat, fee := calculateFee(rawTx)

		tx := &datastore.TransactionData{
			Block:       bd.Block.Height,
			BlockHash:   bd.Block.Hash,
			Confirmed:   true,
			IsCoinbase:  rawTx.Vin[0].IsCoinBase(),
			Transaction: rawTx,
			Fee:         fee,
			FeeSat:      feeSat,
		}

		datastore.AutoBulk.StoreTransaction(tx)
		if len(tx.Transaction.FloData) != 0 {
			events.Publish("flo:floData", tx.Transaction.FloData, tx)
		}
	}
	recentBlocks.Push(&bd)
	return bd, nil
}

func calculateFee(tx *flojson.TxRawResult) (*int64, *float64) {
	if len(tx.Vin) == 0 || tx.Vin[0].IsCoinBase() {
		return nil, nil
	}

	var totalIn float64 = 0

	for i := range tx.Vin {
		hash, err := chainhash.NewHashFromStr(tx.Vin[i].Txid)
		if err != nil {
			log.Error("unable to decode vin txid", tx.Vin[i].Txid, err)
			return nil, nil
		}
		inputTx, err := flo.GetTxVerbose(hash)
		if err != nil {
			log.Error("unable to fetch tx", tx.Vin[i].Txid)
			return nil, nil
		}

		totalIn += inputTx.Vout[tx.Vin[i].Vout].Value
	}

	var totalOut float64 = 0

	for i := range tx.Vout {
		totalOut += tx.Vout[i].Value
	}

	feeSat := int64((totalIn - totalOut) * floutil.SatoshiPerBitcoin)
	fee := float64(feeSat) / floutil.SatoshiPerBitcoin
	return &feeSat, &fee
}
