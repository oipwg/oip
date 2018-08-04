package oip041

import (
	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"github.com/json-iterator/go"
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

	any := jsoniter.Get([]byte(floData))

	el, err := validateOip041(any, tx)
	if err != nil {
		log.Error("validate oip041 failed", logger.Attrs{"err": err})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index("oip041").Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticOip041 struct {
	Artifact jsoniter.Any `json:"artifact"`
	Meta     OMeta        `json:"meta"`
}
type OMeta struct {
	Block       int64                     `json:"block"`
	BlockHash   string                    `json:"block_hash"`
	Deactivated bool                      `json:"deactivated"`
	Signature   string                    `json:"signature"`
	Txid        string                    `json:"txid"`
	Time        int64                     `json:"time"`
	Tx          datastore.TransactionData `json:"tx"`
}

func validateOip041(any jsoniter.Any, tx datastore.TransactionData) (elasticOip041, error) {
	var el elasticOip041

	o41 := any.Get("oip-041")
	sig := any.Get("signature")

	art := o41.Get("artifact")
	if art.LastError() != nil {
		return el, errors.Wrap(art.LastError(), "oip-041.artifact")
	}

	if len(art.Get("info", "title").ToString()) == 0 {
		return el, errors.New("artifact.info.title missing")
	}

	ok, err := flo.CheckAddress(art.Get("floAddress").ToString(), false)
	if !ok {
		return el, errors.Wrap(err, "invalid FLO address")
	}

	el.Artifact = art
	el.Meta = OMeta{
		Time:        tx.Transaction.Time,
		Txid:        tx.Transaction.Txid,
		Signature:   sig.ToString(),
		BlockHash:   tx.BlockHash,
		Block:       tx.Block,
		Deactivated: false,
		Tx:          tx,
	}

	return el, nil
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
                  "type": "object",
                  "enabled": false
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
                  "dynamic": "true",
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
            "deactivated": {
              "type": "boolean"
            },
            "signature": {
              "type": "text",
              "index": false
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
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
