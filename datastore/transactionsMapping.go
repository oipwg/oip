package datastore

const transactionsMapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "block": {
          "type": "long"
        },
        "block_hash": {
          "type": "keyword",
          "ignore_above": 64
        },
        "tx": {
          "properties": {
            "blockhash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "blocktime": {
              "type": "date",
              "format": "epoch_second"
            },
            "confirmations": {
              "type": "long"
            },
            "floData": {
              "type": "text",
              "fields": {
                "keyword": {
                  "type": "keyword",
                  "ignore_above": 256
                }
              }
            },
            "hash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "hex": {
              "type": "text",
              "index": false
            },
            "locktime": {
              "type": "long"
            },
            "size": {
              "type": "long"
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
            },
            "txid": {
              "type": "keyword",
              "ignore_above": 64
            },
            "version": {
              "type": "long"
            },
            "vin": {
              "type": "object",
              "enabled": false,
              "properties": {
                "coinbase": {
                  "type": "text",
                  "index": false
                },
                "scriptSig": {
                  "properties": {
                    "asm": {
                      "type": "text",
                      "index": false
                    },
                    "hex": {
                      "type": "text",
                      "index": false
                    }
                  }
                },
                "sequence": {
                  "type": "long"
                },
                "txid": {
                  "type": "text",
                  "index": false
                },
                "vout": {
                  "type": "long"
                }
              }
            },
            "vout": {
              "type": "object",
              "enabled": false,
              "properties": {
                "n": {
                  "type": "long"
                },
                "scriptPubKey": {
                  "properties": {
                    "addresses": {
                      "type": "text",
                      "index": false
                    },
                    "asm": {
                      "type": "text",
                      "index": false
                    },
                    "hex": {
                      "type": "text",
                      "index": false
                    },
                    "reqSigs": {
                      "type": "long"
                    },
                    "type": {
                      "type": "text",
                      "index": false
                    }
                  }
                },
                "value": {
                  "type": "long"
                }
              }
            },
            "vsize": {
              "type": "long"
            }
          }
        }
      }
    }
  }
}`
