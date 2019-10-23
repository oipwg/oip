package publishing

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"

	"github.com/bitspill/flod/chaincfg"
	"github.com/bitspill/flod/chaincfg/chainhash"
	"github.com/bitspill/flod/flojson"
	"github.com/bitspill/flod/rpcclient"
	"github.com/bitspill/flod/wire"
	"github.com/bitspill/flosig"
	"github.com/bitspill/floutil"
	"github.com/golang/protobuf/proto"

	"github.com/oipwg/oip/modules/oip"
	"github.com/oipwg/oip/modules/oip5"
)

const MaxFloDataLen = 1040
const AncestorLimit = 1250

type Publisher interface {
	UpdateUtxoSet() error
	Publish(o5 ...*oip5.OipFive) (*PublishResult, error)
	CreateAndSignTx(floData []byte) (*wire.MsgTx, error)
	SendToBlockchain(floData []byte) (*SendToBlockchainResult, error)
	SendToBlockchainMultipart(floData []byte) (*SendToBlockchainResult, error)
	SetTxFee(fee floutil.Amount)
}

// assert Address implements Publisher interface
var _ Publisher = &Address{}

type Address struct {
	fee       floutil.Amount
	addr      floutil.Address
	addrBytes []byte
	wif       floutil.WIF
	keys      map[string]*floutil.WIF

	client *rpcclient.Client
	params *chaincfg.Params

	utxoLock                *sync.Mutex
	utxo                    map[string]*Utxo
	unconfirmed             map[string][]*Utxo
	ancestorToDescendantMap map[string]string
	descendantToAncestorMap map[string]string
}

func NewAddress(client *rpcclient.Client, addr floutil.Address, wif *floutil.WIF, net *chaincfg.Params) (Publisher, error) {
	if addr == nil {
		return nil, errors.New("nil address provided")
	}
	if wif == nil {
		return nil, errors.New("nil wif provided")
	}

	pkh := floutil.Hash160(wif.SerializePubKey())
	// ignore error since only possibility is pkh not being 20 bytes
	// but we know it's a Hash160 from previous line
	a, _ := floutil.NewAddressPubKeyHash(pkh, net)
	if a.EncodeAddress() != addr.EncodeAddress() {
		return nil, errors.New("wif does not match address")
	}

	newPub := &Address{
		addr:      a,
		addrBytes: []byte(a.EncodeAddress()),
		wif:       *wif,
		keys:      make(map[string]*floutil.WIF),
		fee:       0.0001 * floutil.SatoshiPerBitcoin,
		client:    client,
		params:    &chaincfg.MainNetParams,
		utxoLock:  new(sync.Mutex),
	}

	newPub.resetUtxo()
	newPub.keys[string(newPub.addrBytes)] = wif

	return newPub, nil
}

func (a *Address) SetTxFee(fee floutil.Amount) {
	a.fee = fee
}

func (a *Address) SendToBlockchain(floData []byte) (*SendToBlockchainResult, error) {
	if a.client == nil {
		return nil, errors.New("nil rpc client")
	}

	if len(floData) > MaxFloDataLen {
		return nil, errors.New("maximum flo data length exceeded, send as multipart")
	}

	a.utxoLock.Lock()
	defer a.utxoLock.Unlock()

	result, err := a.lockedSendToBlockchain(floData)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Requires lock
func (a *Address) lockedSendToBlockchain(floData []byte) (*SendToBlockchainResult, error) {
	tx, err := a.CreateAndSignTx(floData)
	if err != nil {
		return nil, err
	}
	result, err := a.PushTx(tx)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Requires lock
func (a *Address) PushTx(tx *wire.MsgTx) (*SendToBlockchainResult, error) {
	// panic("no")
	_, err := a.client.SendRawTransaction(tx, false)
	if err != nil {
		return nil, err
	}
	txHash := tx.TxHash()
	result := &SendToBlockchainResult{
		Tx:     []*wire.MsgTx{tx},
		TxHash: []*chainhash.Hash{&txHash},
	}
	if len(tx.TxOut) != 1 {
		a.utxo = make(map[string]*Utxo)
		return result, errors.New("unexpected txOut")
	}
	newUtxo := &Utxo{
		Hash:     &txHash,
		Index:    0,
		PkScript: tx.TxOut[0].PkScript,
		Value:    floutil.Amount(tx.TxOut[0].Value),
		Conf:     0,
	}

	selfKey := keyFromInternalHash(&txHash)
	a.utxo[selfKey] = newUtxo

	for _, in := range tx.TxIn {
		parentKey := keyFromTxIn(in)

		delete(a.utxo, parentKey)
		a.linkAncestors(parentKey, selfKey, newUtxo)
	}
	return result, nil
}

// Requires lock
func (a *Address) CreateAndSignTx(floData []byte) (*wire.MsgTx, error) {
	vin, vout, err := a.buildVinVout(a.fee)
	if err != nil {
		return nil, err
	}

	// ToDo: refactor flosig and flod to process flodata as byte slices instead of string
	tx, err := flosig.CreateAndSignTx(vin, vout, a.keys, a.params, string(floData))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// requires lock on utxo
func (a *Address) buildVinVout(fee floutil.Amount) ([]flosig.Vin, []flosig.Vout, error) {
	total := floutil.Amount(0)
	var vin []flosig.Vin
	for k, u := range a.utxo {
		// check ancestor limit
		if u.Conf == 0 {
			ancestor := a.descendantToAncestorMap[k]
			if len(a.unconfirmed[ancestor]) >= AncestorLimit {
				continue
			}
		}

		vin = append(vin, flosig.Vin{
			Hash:     u.Hash,
			Index:    u.Index,
			PkScript: u.PkScript,
		})
		total += u.Value

		if total >= 2*fee {
			break
		}
	}
	change := total - fee
	if change < fee {
		return nil, nil, errors.New("insufficient balance available")
	}
	vout := []flosig.Vout{{
		Addr:   a.addr,
		Amount: change,
	}}
	return vin, vout, nil
}

func (a *Address) SendToBlockchainMultipart(floData []byte) (*SendToBlockchainResult, error) {
	if len(floData) <= MaxFloDataLen {
		return a.SendToBlockchain(floData)
	}

	a.utxoLock.Lock()
	defer a.utxoLock.Unlock()

	mp := NewMultiPart(a, a.addr, *a.keys[a.addr.EncodeAddress()])
	mp.SetFloData(floData)
	tx0, parts, err := mp.Build()
	if err != nil {
		return nil, err
	}

	txHash := tx0.TxHash()
	result := &SendToBlockchainResult{
		Tx:     []*wire.MsgTx{tx0},
		TxHash: []*chainhash.Hash{&txHash},
	}

	_, err = a.PushTx(tx0)
	if err != nil {
		return result, err
	}

	for i := 1; i < len(parts); i++ {
		res, err := a.lockedSendToBlockchain(parts[i])
		if res != nil {
			result.Tx = append(result.Tx, res.Tx...)
			result.TxHash = append(result.TxHash, res.TxHash...)
		}
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

type SendToBlockchainResult struct {
	Tx     []*wire.MsgTx
	TxHash []*chainhash.Hash
}

type PublishResult struct {
	Sbr []*SendToBlockchainResult
}

func (a *Address) Publish(o5 ...*oip5.OipFive) (*PublishResult, error) {
	result := &PublishResult{}
	for _, record := range o5 {
		serRecord, err := a.genO5SerializedSignedMessage(record)
		if err != nil {
			return result, err
		}

		enc := base64.StdEncoding
		floData := make([]byte, 4+enc.EncodedLen(len(serRecord)))
		copy(floData, "p64:")
		enc.Encode(floData[4:], serRecord)

		recordSendResult, err := a.SendToBlockchainMultipart(floData)
		if recordSendResult != nil {
			result.Sbr = append(result.Sbr, recordSendResult)
		}
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func (a *Address) genO5SerializedSignedMessage(o5 *oip5.OipFive) ([]byte, error) {
	serializedProtoMessage, err := proto.Marshal(o5)
	if err != nil {
		return nil, err
	}
	spm64 := base64.StdEncoding.EncodeToString(serializedProtoMessage)
	sig64, err := flosig.SignMessagePk(spm64, "Florincoin", a.wif.PrivKey, a.wif.CompressPubKey)
	if err != nil {
		return nil, err
	}
	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil {
		return nil, err
	}
	msg := &oip.SignedMessage{
		SerializedMessage: serializedProtoMessage,
		MessageType:       oip.MessageTypes_Multipart,
		SignatureType:     oip.SignatureTypes_Flo,
		PubKey:            a.addrBytes,
		Signature:         sig,
	}
	serializedSignedMessage, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return serializedSignedMessage, nil
}

func (a *Address) UpdateUtxoSet() error {
	if a.client == nil {
		return errors.New("nil rpc client")
	}

	a.utxoLock.Lock()
	defer a.utxoLock.Unlock()

	skip := 0
	resultsPerRequest := 50000

	utxo := make(map[string]*fastUtxo)
	mempoolUtxo := make(map[string]*fastUtxo)

requestMore:
	res, err := a.client.SearchRawTransactionsVerbose(a.addr,
		skip, resultsPerRequest, false,
		false, []string{a.addr.String()})
	if err != nil && !isErrRPCNoTxInfo(err) {
		return err
	}

	// res contains all transactions associated with the controlled address
	// vin/vout will be filtered by the rpc server

	for _, tx := range res {
		// Add vout to utxo set clearing spent along the way
		for i := range tx.Vout {
			vout := &tx.Vout[i]
			k := keyFromTxVout(tx, vout)
			u := fastUtxo{
				Conf:     tx.Confirmations,
				Hash:     &tx.Txid,
				Index:    vout.N,
				PkScript: &vout.ScriptPubKey.Hex,
				Value:    vout.Value,
			}
			utxo[k] = &u

			// Unconfirmed transactions will be unordered so must perform additional tracking
			if tx.Confirmations == 0 {
				u.VinPrevOut = tx.Vin
				mempoolUtxo[k] = &u
			}
		}

		// purge spent vin from utxo set
		for i := range tx.Vin {
			k := keyFromVinPrevOut(&tx.Vin[i])
			delete(utxo, k)
			delete(mempoolUtxo, k)
		}
	}

	// request subsequent pages of tx
	if len(res) != 0 {
		skip += len(res)
		goto requestMore
	}
	// cleanup utxo of any out of order mempool spends
	for _, v := range mempoolUtxo {
		for i := range v.VinPrevOut {
			prev := &v.VinPrevOut[i]
			k := keyFromVinPrevOut(prev)
			delete(utxo, k)
		}
	}

	err = a.buildAncestralTrees(utxo)
	if err != nil {
		return err
	}

	return nil
}

func (a *Address) buildAncestralTrees(utxo map[string]*fastUtxo) error {
	// convert fastUtxo to a usable UTXO set
	a.utxo = make(map[string]*Utxo)
	a.unconfirmed = make(map[string][]*Utxo)
	for selfKey, v := range utxo {
		pks, err := hex.DecodeString(*v.PkScript)
		if err != nil {
			a.resetUtxo()
			return err
		}
		amt, err := floutil.NewAmount(v.Value)
		if err != nil {
			a.resetUtxo()
			return err
		}
		hash, err := chainhash.NewHashFromStr(*v.Hash)
		if err != nil {
			a.resetUtxo()
			return err
		}
		u := &Utxo{
			Hash:     hash,
			Index:    v.Index,
			PkScript: pks,
			Value:    amt,
			Conf:     v.Conf,
		}

		a.utxo[selfKey] = u

		// unconfirmed, determine ancestor/descendant
		if v.Conf == 0 {
			for i := range v.VinPrevOut {
				parentKey := keyFromVinPrevOut(&v.VinPrevOut[i])
				a.linkAncestors(parentKey, selfKey, u)
			}
		}
	}
	return nil
}

func (a *Address) resetUtxo() {
	a.utxo = make(map[string]*Utxo)
	a.unconfirmed = make(map[string][]*Utxo)
	a.ancestorToDescendantMap = make(map[string]string)
	a.descendantToAncestorMap = make(map[string]string)
}

func (a *Address) linkAncestors(parentKey string, selfKey string, u *Utxo) {
	ancestorKey, hasAncestor := a.descendantToAncestorMap[parentKey]
	descendantKey, hasDescendant := a.ancestorToDescendantMap[selfKey]

	if !hasAncestor {
		ancestorKey = selfKey
	}
	// make initial connections
	a.ancestorToDescendantMap[parentKey] = selfKey
	a.ancestorToDescendantMap[ancestorKey] = selfKey
	a.descendantToAncestorMap[selfKey] = ancestorKey
	// add self to ancestor descendants chain
	a.unconfirmed[ancestorKey] = append(a.unconfirmed[ancestorKey], u)
	// if have descendant connect to common ancestor
	if hasDescendant {
		a.descendantToAncestorMap[descendantKey] = ancestorKey
		a.ancestorToDescendantMap[ancestorKey] = descendantKey

		if descendantKey != ancestorKey {
			// move own descendants to common ancestor chain
			a.unconfirmed[ancestorKey] = append(a.unconfirmed[ancestorKey], a.unconfirmed[descendantKey]...)
			for i := range a.unconfirmed[ancestorKey] {
				k := keyFromInternalHash(a.unconfirmed[ancestorKey][i].Hash)
				a.descendantToAncestorMap[k] = ancestorKey
			}
			delete(a.unconfirmed, descendantKey)
		}
	}
}

func isErrRPCNoTxInfo(err error) bool {
	re, ok := err.(*flojson.RPCError)
	return ok && re.Code == flojson.ErrRPCNoTxInfo
}

type Utxo struct {
	Hash     *chainhash.Hash
	Index    uint32
	PkScript []byte
	Value    floutil.Amount
	Conf     uint64
}

type fastUtxo struct {
	Conf       uint64
	Hash       *string
	Index      uint32
	PkScript   *string
	Value      float64
	VinPrevOut []flojson.VinPrevOut
}

func keyFromTxVout(tx *flojson.SearchRawTransactionsResult, vout *flojson.Vout) string {
	return tx.Txid + "." + strconv.Itoa(int(vout.N))
}

func keyFromVinPrevOut(vpo *flojson.VinPrevOut) string {
	return vpo.Txid + "." + strconv.Itoa(int(vpo.Vout))
}

func keyFromTxIn(tin *wire.TxIn) string {
	return tin.PreviousOutPoint.Hash.String() + "." + strconv.Itoa(int(tin.PreviousOutPoint.Index))
}

func keyFromInternalHash(hash *chainhash.Hash) string {
	return hash.String() + ".0"
}
