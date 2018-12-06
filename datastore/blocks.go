package datastore

import (
	"context"
	"encoding/json"

	"github.com/bitspill/flod/flojson"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	RegisterMapping("blocks", "blocks.json")
}

func GetLastBlock(ctx context.Context) (BlockData, error) {
	sRes, err := client.Search().
		Index(Index("blocks")).
		Sort("block.height", false).
		Size(1).
		Do(ctx)

	if err != nil {
		return BlockData{}, err
	}

	log.Info("GetLastBlockHeight query time (ms) ", sRes.TookInMillis)

	if sRes.TotalHits() == 0 {
		return BlockData{}, nil
	}

	var br BlockData
	err = json.Unmarshal(*sRes.Hits.Hits[0].Source, &br)

	if err != nil {
		return BlockData{}, err
	}

	return br, nil
}

func StoreBlock(ctx context.Context, b BlockData) (*elastic.IndexResponse, error) {
	put1, err := client.Index().
		Index(Index("blocks")).
		Type("_doc").
		Id(b.Block.Hash).
		BodyJson(b).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return put1, nil
}

func GetBlockFromID(ctx context.Context, id string) (BlockData, error) {
	get, err := client.Get().Index(Index("blocks")).Type("_doc").Id(id).Do(ctx)
	if err != nil {
		return BlockData{}, err
	}
	if get.Found {
		var bd BlockData
		err := json.Unmarshal(*get.Source, &bd)
		return bd, err
	} else {
		return BlockData{}, errors.New("ID not found")
	}
}

type BlockData struct {
	Block             *flojson.GetBlockVerboseResult `json:"block"`
	SecSinceLastBlock int64                          `json:"sec_since_last_block"`
}
