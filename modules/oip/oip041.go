package oip

import (
	"encoding/json"

	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	log.Info("init oip41")
	events.Bus.SubscribeAsync("modules:oip:oip041", on41, false)

	datastore.RegisterMapping("oip041", oip041Mapping)
}

func on41(floData string, tx datastore.TransactionData) {
	log.Info("oip041 ", tx.Transaction.Txid)

	var tl topLevel
	err := json.Unmarshal([]byte(floData), &tl)
	if err != nil {
		return
	}

	art, err := validateOip041(tl)
	if err != nil {
		spew.Dump(err)
		return
	}

	el := elasticOip041{
		Artifact: tl.Oip041.Artifact,
		Meta: OMeta{
			Block:     tx.Block,
			BlockHash: tx.BlockHash,
			Txid:      tx.Transaction.Txid,
			Tx:        tx,
			Time:      art.Timestamp,
			Signature: tl.Oip041.Signature,
		},
	}

	bir := elastic.NewBulkIndexRequest().Index("oip041").Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type topLevel struct {
	Oip041 O41 `json:"oip-041"`
}

type O41 struct {
	Artifact  *json.RawMessage `json:"artifact"`
	Signature string           `json:"signature"`
}

type Artifact struct {
	Info       *Info  `json:"info"`
	FloAddress string `json:"publisher"`
	Timestamp  int64  `json:"timestamp"`
}

type Info struct {
	Title string `json:"title"`
}

type elasticOip041 struct {
	Artifact *json.RawMessage `json:"artifact"`
	Meta     OMeta            `json:"meta"`
}
type OMeta struct {
	Block     int64                     `json:"block"`
	BlockHash string                    `json:"block_hash"`
	Signature string                    `json:"signature"`
	Txid      string                    `json:"txid"`
	Time      int64                     `json:"time"`
	Tx        datastore.TransactionData `json:"tx"`
}

func validateOip041(tl topLevel) (Artifact, error) {
	var art Artifact

	if tl.Oip041.Artifact != nil && len(*tl.Oip041.Artifact) != 0 {
		err := json.Unmarshal(*tl.Oip041.Artifact, &art)
		if err != nil {
			return art, errors.Wrap(err, "artifact invalid json")
		}

		if art.Info == nil || len(art.Info.Title) == 0 {
			return art, errors.New("artifact.info.title missing")
		}

		ok, err := flo.CheckAddress(art.FloAddress, false)
		if !ok {
			return art, errors.Wrap(err, "invalid FLO address")
		}
	} else {
		return art, errors.New("no Artifact data")
	}

	return art, nil
}

const oip041Mapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "artifact": {
          "properties": {
            "floAddress": {
              "type": "keyword",
              "ignore_above": 36
            },
            "info": {
              "properties": {
                "description": {
                  "type": "text",
                  "fields": {
                    "keyword": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "extraInfo": {
                  "dynamic": "true",
                  "properties": {
                    "ISRC": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "artist": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "company": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "composers": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "copyright": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "coverArt": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "creator": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "director": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "distributor": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "episodeNum": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "episodeTitle": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "extID": {
                      "properties": {
                        "localID": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "namespace": {
                          "type": "keyword",
                          "ignore_above": 256
                        }
                      }
                    },
                    "genre": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "gis": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "gpsBounds": {
                      "type": "object",
                      "enabled": false
                    },
                    "grsBounds": {
                      "type": "object",
                      "enabled": false
                    },
                    "partyRole": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "partyType": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "posterFrame": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "preview": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "seasonNum": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "tags": {
                      "type": "text"
                    },
                    "year": {
                      "type": "date",
                      "format": "year"
                    }
                  }
                },
                "gpsBounds": {
                  "type": "object",
                  "enabled": false
                },
                "nsfw": {
                  "type": "boolean"
                },
                "title": {
                  "type": "text",
                  "fields": {
                    "keyword": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "year": {
                  "type": "date",
                  "format": "year"
                }
              }
            },
            "payment": {
              "properties": {
                "addresses": {
                  "properties": {
                    "address": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "token": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "disPer": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "fiat": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "maxdisc": {
                  "type": "long"
                },
                "promoter": {
                  "type": "long"
                },
                "retailer": {
                  "type": "long"
                },
                "scale": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "sugTip": {
                  "type": "long"
                },
                "tokens": {
                  "properties": {
                    "BTC": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "HBCUCOIN": {
                      "type": "long"
                    },
                    "KEEPTHEFAITH": {
                      "type": "long"
                    },
                    "LTBCOIN": {
                      "type": "long"
                    },
                    "TATIANAFAN": {
                      "type": "long"
                    },
                    "btc": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                }
              }
            },
            "publisher": {
              "type": "keyword",
              "ignore_above": 36
            },
            "storage": {
              "properties": {
                "files": {
                  "properties": {
                    "disBuy": {
                      "type": "boolean"
                    },
                    "disPlay": {
                      "type": "boolean"
                    },
                    "disallowBuy": {
                      "type": "boolean"
                    },
                    "disallowPlay": {
                      "type": "boolean"
                    },
                    "dissallowBuy": {
                      "type": "boolean"
                    },
                    "dname": {
                      "type": "text",
                      "fields": {
                        "keyword": {
                          "type": "keyword",
                          "ignore_above": 256
                        }
                      }
                    },
                    "duration": {
                      "type": "long"
                    },
                    "fName": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "fSize": {
                      "type": "long"
                    },
                    "fame": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "fname": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "fsize": {
                      "type": "long"
                    },
                    "minBuy": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "minPlay": {
                      "type": "long"
                    },
                    "promo": {
                      "type": "long"
                    },
                    "retail": {
                      "type": "long"
                    },
                    "subtype": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "sugBuy": {
                      "type": "long"
                    },
                    "sugPlay": {
                      "type": "long"
                    },
                    "tokenlyID": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "type": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "location": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "network": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "signature": {
                  "type": "text",
                  "index": false
                }
              }
            },
            "subtype": {
              "type": "keyword",
              "ignore_above": 256
            },
            "timestamp": {
              "type": "date",
              "format": "epoch_second"
            },
            "type": {
              "type": "keyword",
              "ignore_above": 256
            }
          }
        },
        "meta": {
          "properties": {
            "block": {
              "type": "long"
            },
            "block_hash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "signature": {
              "type": "text",
              "index": false
            },
            "time": {
              "type": "long"
            },
            "tx": {
              "type": "object",
              "enabled": false
            },
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            }
          }
        }
      }
    }
  }
}
`
