package datastore

const blocksMapping = `{
  "settings": {
    "number_of_shards": 2
  },
  "mappings": {
    "_doc": {
      "dynamic": "strict",
      "properties": {
        "block": {
          "properties": {
            "bits": {
              "type": "keyword",
              "ignore_above": 8
            },
            "confirmations": {
              "type": "long"
            },
            "difficulty": {
              "type": "float"
            },
            "hash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "height": {
              "type": "long"
            },
            "merkleroot": {
              "type": "keyword",
              "ignore_above": 64
            },
            "nextblockhash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "nonce": {
              "type": "long"
            },
            "previousblockhash": {
              "type": "keyword",
              "ignore_above": 64
            },
            "rawtx": {
              "type": "object",
              "enabled": false
            },
            "size": {
              "type": "long"
            },
            "strippedsize": {
              "type": "long"
            },
            "time": {
              "type": "date",
              "format": "epoch_second"
            },
            "version": {
              "type": "long"
            },
            "versionHex": {
              "type": "keyword",
              "ignore_above": 8
            },
            "weight": {
              "type": "long"
            }
          }
        },
        "sec_since_last_block": {
          "type": "long"
        }
      }
    }
  }
}`
