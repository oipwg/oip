package oip041

import (
	"context"
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/flo"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"gopkg.in/olivere/elastic.v6"
)

const oip41IndexName = "oip041"

var artRouter = httpapi.NewSubRoute("/oip041/artifact")

func init() {
	log.Info("init oip41")
	events.Bus.SubscribeAsync("modules:oip:oip041", on41, false)

	datastore.RegisterMapping(oip41IndexName, oip041Mapping)

	artRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest).Queries("nsfw", "{nsfw}")
	artRouter.HandleFunc("/get/latest/{limit:[0-9]+}", handleLatest)
	artRouter.HandleFunc("/get/{id:[a-f0-9]+}", handleGet)
}

func handleLatest(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	size, _ := strconv.ParseInt(opts["limit"], 10, 0)
	if size <= 0 || size > 1000 {
		size = -1
	}

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
	)

	if n, ok := opts["nsfw"]; ok {
		nsfw, _ := strconv.ParseBool(n)
		q.Must(elastic.NewTermQuery("artifact.info.nsfw", nsfw))
		log.Info("nsfw: %t", nsfw)
	}

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(datastore.Index(oip41IndexName)).
		Type("_doc").
		Query(q).
		Size(int(size)).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		httpapi.RespondJSON(w, 500, map[string]interface{}{
			"error": "database error",
		})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"count":   len(results.Hits.Hits),
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	var opts = mux.Vars(r)

	q := elastic.NewBoolQuery().Must(
		elastic.NewTermQuery("meta.deactivated", false),
		// elastic.NewTermQuery("_id", opts["id"]),
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(datastore.Index(oip41IndexName)).
		Type("_doc").
		Query(q).
		Size(1).
		Sort("meta.time", false).
		FetchSourceContext(fsc).
		Do(context.TODO())

	if err != nil {
		log.Error("elastic search failed", logger.Attrs{"err": err})
		return
	}

	sources := make([]interface{}, len(results.Hits.Hits))
	for k, v := range results.Hits.Hits {
		sources[k] = v.Source
	}

	httpapi.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func on41(floData string, tx *datastore.TransactionData) {
	log.Info("oip041 ", tx.Transaction.Txid)

	any := jsoniter.Get([]byte(floData))

	el, err := validateOip041(any, tx)
	if err != nil {
		log.Error("validate oip041 failed", logger.Attrs{"err": err})
		return
	}

	bir := elastic.NewBulkIndexRequest().Index(datastore.Index("oip041")).Type("_doc").Id(tx.Transaction.Txid).Doc(el)
	datastore.AutoBulk.Add(bir)
}

type elasticOip041 struct {
	Artifact interface{} `json:"artifact"`
	Meta     OMeta       `json:"meta"`
}
type OMeta struct {
	Block       int64                      `json:"block"`
	BlockHash   string                     `json:"block_hash"`
	Deactivated bool                       `json:"deactivated"`
	Signature   string                     `json:"signature"`
	Time        int64                      `json:"time"`
	Tx          *datastore.TransactionData `json:"tx"`
	Txid        string                     `json:"txid"`
	Type        string                     `json:"type"`
}

func validateOip041(any jsoniter.Any, tx *datastore.TransactionData) (elasticOip041, error) {
	var el elasticOip041

	o41 := any.Get("oip-041")
	sig := o41.Get("signature")

	art := o41.Get("artifact")
	if art.LastError() != nil {
		return el, errors.Wrap(art.LastError(), "oip-041.artifact")
	}

	if len(art.Get("info", "title").ToString()) == 0 {
		return el, errors.New("artifact.info.title missing")
	}

	ok, err := flo.CheckAddress(art.Get("publisher").ToString())
	if !ok {
		return el, errors.Wrap(err, "invalid FLO address")
	}

	el.Artifact = art.GetInterface()
	el.Meta = OMeta{
		Block:       tx.Block,
		BlockHash:   tx.BlockHash,
		Deactivated: false,
		Signature:   sig.ToString(),
		Time:        tx.Transaction.Time,
		Tx:          tx,
		Txid:        tx.Transaction.Txid,
		Type:        "oip041",
	}

	return el, nil
}

const oip041Mapping = `{
  "settings": {
    "number_of_shards": 2
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
            },
            "type": {
              "type": "keyword",
              "ignore_above": 16
            }
          }
        }
      }
    }
  }
}
`
