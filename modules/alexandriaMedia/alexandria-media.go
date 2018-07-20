package alexandriaMedia

import (
	"github.com/azer/logger"
	"github.com/bitspill/oip/datastore"
	"github.com/bitspill/oip/events"
	"github.com/json-iterator/go"
	"gopkg.in/olivere/elastic.v6"
)

func init() {
	log.Info("init alexandria-media")
	events.Bus.SubscribeAsync("modules:oip:alexandriaMedia", onAlexandriaMedia, false)
	datastore.RegisterMapping("alexandria-media", amMapping)
}

func onAlexandriaMedia(floData string, tx datastore.TransactionData) {
	log.Info("onAlexandriaMedia", logger.Attrs{"txid": tx.Transaction.Txid})

	bytesFloData := []byte(floData)

	am := jsoniter.Get(bytesFloData, "alexandria-media")
	title := am.Get("info", "title").ToString()
	if len(title) != 0 {
		bir := elastic.NewBulkIndexRequest().Index("alexandria-media").Type("_doc").Doc(am.GetInterface()).Id(tx.Transaction.Txid)
		datastore.AutoBulk.Add(bir)
	} else {
		log.Info("no title", logger.Attrs{"txid": tx.Transaction.Txid})
	}
}

const amMapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
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
    }
  }
}
`
