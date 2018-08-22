package oip042

const publishOip042ArtifactMapping = `{
  "settings": {
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "artifact": {
          "properties": {
            "details": {
              "properties": {
                "NBCItaxID": {
                  "type": "long"
                },
                "NCBItaxID": {
                  "type": "long"
                },
                "artNotes": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "artist": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "date": {
                  "type": "long"
                },
                "defocus": {
                  "type": "long"
                },
                "dosage": {
                  "type": "long"
                },
                "genre": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "institution": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "lab": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "magnification": {
                  "type": "long"
                },
                "microscopist": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "roles": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "scopeName": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "sid": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "speciesName": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "strain": {
                  "type": "keyword",
                  "ignore_above": 256
                },
                "tiltConstant": {
                  "type": "long"
                },
                "tiltMax": {
                  "type": "long"
                },
                "tiltMin": {
                  "type": "long"
                },
                "tiltSingleDual": {
                  "type": "long"
                },
                "tiltStep": {
                  "type": "long"
                }
              }
            },
            "floAddress": {
              "type": "keyword",
              "ignore_above": 256
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
                "tags": {
                  "type": "text",
                  "fields": {
                    "keyword": {
                      "type": "keyword",
                      "ignore_above": 256
                    }
                  }
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
                }
              }
            },
            "signature": {
              "type": "text",
              "index": false
            },
            "storage": {
              "properties": {
                "files": {
                  "properties": {
                    "cType": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "dname": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "fNotes": {
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
                    "software": {
                      "type": "keyword",
                      "ignore_above": 256
                    },
                    "subtype": {
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
                }
              }
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
