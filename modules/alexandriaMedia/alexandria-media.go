package alexandriaMedia

import (
	"context"
	"net/http"
	"strconv"

	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/bitspill/oip/httpapi"
	"github.com/gorilla/mux"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
	"time"
)

const amIndexName = "alexandria-media"

var artRouter = httpapi.NewSubRoute("/alexandria/artifact")

func init() {
	log.Info("init alexandria-media")
	events.Bus.SubscribeAsync("modules:oip:alexandriaMedia", onAlexandriaMedia, false)
	datastore.RegisterMapping(amIndexName, amMapping)
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

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(amIndexName).
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
		elastic.NewPrefixQuery("meta.txid", opts["id"]),
	)

	fsc := elastic.NewFetchSourceContext(true).
		Include("artifact.*", "meta.block_hash", "meta.txid", "meta.block", "meta.time")

	results, err := datastore.Client().
		Search(amIndexName).
		Type("_doc").
		Query(q).
		Size(1).
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
		"total":   results.Hits.TotalHits,
		"results": sources,
	})
}

func onAlexandriaMedia(floData string, tx datastore.TransactionData) {
	log.Info("onAlexandriaMedia", logger.Attrs{"txid": tx.Transaction.Txid})

	bytesFloData := []byte(floData)
	a := jsoniter.Get(bytesFloData)
	am := a.Get("alexandria-media")
	title := am.Get("info", "title").ToString()
	artTime := am.Get("timestamp").ToInt64()
	if artTime > 999999999999 {
		// correct ms timestamps to s
		log.Info("Scaling timestamp", logger.Attrs{"txid": tx.Transaction.Txid, "artTime": artTime, "time": time.Unix(artTime, 0)})
		artTime /= 1000
	}
	if len(title) != 0 {
		el := elasticAm{
			Artifact: am.GetInterface(),
			Meta: AmMeta{
				Block:       tx.Block,
				BlockHash:   tx.BlockHash,
				Deactivated: false,
				Signature:   a.Get("signature").ToString(),
				Txid:        tx.Transaction.Txid,
				Tx:          tx,
				Time:        artTime,
				Type:        "alexandria-media",
			},
		}

		bir := elastic.NewBulkIndexRequest().Index(amIndexName).Type("_doc").Doc(el).Id(tx.Transaction.Txid)
		datastore.AutoBulk.Add(bir)
	} else {
		log.Info("no title", logger.Attrs{"txid": tx.Transaction.Txid})
	}
}

type elasticAm struct {
	Artifact interface{} `json:"artifact"`
	Meta     AmMeta      `json:"meta"`
}

type AmMeta struct {
	Block       int64                     `json:"block"`
	BlockHash   string                    `json:"block_hash"`
	Deactivated bool                      `json:"deactivated"`
	Signature   string                    `json:"signature"`
	Txid        string                    `json:"txid"`
	Time        int64                     `json:"time"`
	Tx          datastore.TransactionData `json:"tx"`
	Type        string                    `json:"type"`
}

const amMapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "artifact": {
          "properties": {
            "files": {
              "properties": {
                "dname": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "duration": {
                  "type": "text",
                  "index": false
                },
                "fname": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "minBuy": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "minPlay": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "runtime": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "sugBuy": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "sugPlay": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "type": {
                  "type": "keyword",
                  "ignore_above": 256
                }
              }
            },
            "info": {
              "properties": {
                "artist": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "collection": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "description": {
                  "type": "text"
                },
                "extra-info": {
                  "dynamic": "true",
                  "properties": {
                    "Bitcoin Address": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "DHT Hash": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "RottenTom": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "albumtrack": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "artist": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "collection": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "company": {
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
                    "creators2": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "displayname": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "filename": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "files": {
                      "properties": {
                        "disallowBuy": {
                          "type": "boolean"
                        },
                        "disallowPlay": {
                          "type": "boolean"
                        },
                        "dname": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "duration": {
                          "type": "float"
                        },
                        "fname": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "minBuy": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "minPlay": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "runtime": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "sugBuy": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "sugPlay": {
                          "type": "keyword",
                          "ignore_above": 256
                        },
                        "type": {
                          "type": "keyword",
                          "ignore_above": 256
                        }
                      }
                    },
                    "filetype": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "fkasljdflk": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "genere": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "genre": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "poster": {
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
                    "pwyw": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "runtime": {
                      "type": "long"
                    },
                    "tags": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "track02": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "track03": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "track04": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "trailer": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "wwwId": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "genre": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "payment": {
                  "type": "object"
                },
                "publisher": {
                  "type": "keyword",
                  "ignore_above": 36
                },
                "runtime": {
                  "type": "long"
                },
                "size": {
                  "type": "long"
                },
                "timestamp": {
                  "type": "date",
                  "format": "epoch_second"
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
                "torrent": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "type": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "wwwId": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "year": {
                  "type": "date",
                  "format": "year"
                }
              }
            },
            "payment": {
              "properties": {
                "amount": {
                  "type": "text"
                },
                "currency": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "fiat": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "paymentAddress": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "paymentToken": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "scale": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "tokens": {
                  "properties": {
                    "btc": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
                },
                "type": {
                  "type": "keyword",
                  "ignore_above": 256
                }
              }
            },
            "publisher": {
              "type": "keyword",
              "ignore_above": 256
            },
            "runtime": {
              "type": "long"
            },
            "signature": {
              "type": "text",
              "index": false
            },
            "storage": {
              "properties": {
                "files": {
                  "properties": {
                    "dname": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "duration": {
                      "type": "float"
                    },
                    "fname": {
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
                }
              }
            },
            "timestamp": {
              "type": "date",
              "format": "epoch_second"
            },
            "torrent": {
              "type": "keyword",
              "ignore_above": 256
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
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            },
            "tx": {
              "type": "object",
              "enabled": false
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
