package oip042

import "github.com/bitspill/oip/datastore"

type OMeta struct {
	Block     int64                      `json:"block"`
	BlockHash string                     `json:"block_hash"`
	Completed bool                       `json:"completed"`
	Signature string                     `json:"signature"`
	Txid      string                     `json:"txid"`
	Time      int64                      `json:"time"`
	Tx        *datastore.TransactionData `json:"tx"`
	Type      string                     `json:"type"`
}

type AMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Deactivated bool                       `json:"deactivated"`
	Signature   string                     `json:"signature"`
	Txid        string                     `json:"txid"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"tx"`
	Type        string                     `json:"type"`
}

type elasticOip042Edit struct {
	Edit interface{} `json:"edit"`
	Meta OMeta       `json:"meta"`
}

type elasticOip042Transfer struct {
	Transfer interface{} `json:"transfer"`
	Meta     OMeta       `json:"meta"`
}

type elasticOip042Artifact struct {
	Artifact interface{} `json:"artifact"`
	Meta     AMeta       `json:"meta"`
}

type elasticOip042Pub struct {
	Pub  interface{} `json:"publisher"`
	Meta AMeta       `json:"meta"`
}

type elasticOip042Influencer struct {
	Influencer interface{} `json:"influencer"`
	Meta       AMeta       `json:"meta"`
}

type elasticOip042Platform struct {
	Platform interface{} `json:"platform"`
	Meta     AMeta       `json:"meta"`
}

type elasticOip042Autominer struct {
	Autominer interface{} `json:"autominer"`
	Meta      AMeta       `json:"meta"`
}

type elasticOip042Pool struct {
	Pool interface{} `json:"pool"`
	Meta AMeta       `json:"meta"`
}
