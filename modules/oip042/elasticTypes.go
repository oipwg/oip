package oip042

import "github.com/oipwg/oip/datastore"

type OMeta struct {
	Block        int64                      `json:"block"`
	BlockHash    string                     `json:"block_hash"`
	Completed    bool                       `json:"completed"`
	Defective    bool                       `json:"defective"`
	Signature    string                     `json:"signature"`
	Time         int64                      `json:"time"`
	Tx           *datastore.TransactionData `json:"tx"`
	Txid         string                     `json:"txid"`
	Type         string                     `json:"type"`
	OriginalTxid string                     `json:"originalTxid"`
	PriorTxid    string                     `json:"priorTxid"`
}

type AMeta struct {
	Block         int64                      `json:"block"`
	BlockHash     string                     `json:"block_hash"`
	Deactivated   bool                       `json:"deactivated"`
	Blacklist     Blacklist                  `json:"blacklist"`
	Latest        bool                       `json:"latest"`
	OriginalTxid  string                     `json:"originalTxid"`
	PreviousEdits []string                   `json:"previousEdits"`
	Signature     string                     `json:"signature"`
	Time          int64                      `json:"time"`
	Tx            *datastore.TransactionData `json:"tx"`
	Txid          string                     `json:"txid"`
	Type          string                     `json:"type"`
}

type Blacklist struct {
	Blacklisted bool   `json:"blacklisted"`
	Filter      string `json:"filter"`
}

type elasticOip042Edit struct {
	Edit  interface{} `json:"edit"`
	Meta  OMeta       `json:"meta"`
	Patch string      `json:"patch"`
}

type elasticOip042Transfer struct {
	Transfer interface{} `json:"transfer"`
	Meta     OMeta       `json:"meta"`
}

type elasticOip042Artifact struct {
	Artifact      interface{} `json:"artifact"`
	Meta          AMeta       `json:"meta"`
	LinkedRecords map[int]interface{} `json:"linkedRecords"`
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
