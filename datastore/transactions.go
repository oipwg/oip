package datastore

import (
	"context"
	"encoding/json"

	"github.com/bitspill/flod/flojson"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	RegisterMapping("transactions", "transactions.json")
}

func StoreTransaction(ctx context.Context, t *TransactionData) (*elastic.IndexResponse, error) {
	put1, err := client.Index().
		Index("transaction").
		Type("_doc").
		Id(t.Transaction.Txid).
		BodyJson(t).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return put1, nil
}

func GetTransactionFromID(ctx context.Context, id string) (TransactionData, error) {
	get, err := client.Get().Index(Index("transactions")).Type("_doc").Id(id).Do(ctx)
	if err != nil {
		return TransactionData{}, err
	}
	if get.Found {
		var td TransactionData
		err := json.Unmarshal(*get.Source, &td)
		return td, err
	} else {
		return TransactionData{}, errors.New("ID not found")
	}
}

type TransactionData struct {
	Block       int64                `json:"block"`
	BlockHash   string               `json:"block_hash"`
	Confirmed   bool                 `json:"confirmed"`
	IsCoinbase  bool                 `json:"is_coinbase"`
	Transaction *flojson.TxRawResult `json:"tx"`
}
