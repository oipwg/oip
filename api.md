API
=

- oipd
  - oip/daemon/version
  - oip/floData/search?q={query}
- artifacts (oip41 & oip042)
  - oip/artifact/get/latest?nsfw=true/false
  - oip/artifact/get/{id:[a-f0-9]+}
  - oip/artifact/search?q={query}
- multipart
  - oip/multipart/get/ref/{ref:[a-f0-9]+}
  - oip/multipart/get/id/{id:[a-f0-9]+}
- alexandria-media
  - oip/alexandria/artifact/get/latest
  - oip/alexandria/artifact/get/{id:[a-f0-9]+}
- alexandria-publisher
  - oip/alexandria/publisher/get/latest
  - oip/alexandria/publisher/get/{address:[A-Za-z0-9]+}
- oip041
  - oip/oip041/artifact/get/latest?nsfw=true/false
  - oip/oip041/artifact/get/{id:[a-f0-9]+}
- oip042
  - oip/oip042/artifact/get/latest?nsfw=true/false
  - oip/oip042/record/get/{originalTxid}
  - oip/oip042/record/get/{originalTxid}/version/{editRecordTxid}
  - oip/oip042/edit/get/{editRecordTxid}
  - oip/oip042/edit/search?q={query}
- oip5
  - oip/o5/record/get/latest
  - oip/o5/record/get/{id:[a-f0-9]+}
  - oip/o5/record/mapping/{tmpl:tmpl_[a-fA-F0-9]{8}(?:,tmpl_[a-fA-F0-9]{8})*}
  - oip/o5/record/search?q={query}
  - oip/o5/template/get/latest
  - oip/o5/template/get/{id:[a-fA-F0-9]+}
  - oip/o5/template/search?q={query}

## Common Query Params
All API routes which may return multiple results
also have `after`, `limit`, `page` and `sort` query
parameters which can be used to facilitate pagination

#### After
Following any search the `next` value from the results
may be used as the `after` to continue with the next
page of results  
Used for deep page results.

#### Page
The `page` query parameter may be used to easily jump
to specific results pages however may only be used within
the first 10,000 results - beyond that `after` must be used -
`page` is ignored when `after` provided

#### Limit
Number of results to return per request, range 1-1000; default 10

#### Sort
Allows control over the sort order of results, `sort` is a
string composed of a delimited list of fields and direction

`fieldname:[a|d]$fieldname:[a|d]`
ex: `sort=tx.size:a$tx.time:d`


## Oip5 Examples

#### Latest Records
`http://localhost:1606/oip/o5/record/get/latest?limit=1&page=66`

```json
{
  "count": 1,
  "total": 151,
  "results": [
    {
      "meta": {
        "signed_by": "FEQXuxEgfGoEnZbPCHPfHQDMUqXq6tpw4X",
        "publisher_name": "Devon James",
        "block_hash": "ac501599a79a9616b0ad73d441f04a3e21167301552395df6dc1768124fe32df",
        "txid": "fb049475c743a0486129627013d76a4cc83792bce56f4805e8e18d095dd1d5be",
        "block": 3520788,
        "time": 1561877100,
        "type": "oip5"
      },
      "record": {
        "details": {
          "tmpl_90BD561D": {
            "tweetUrl": "https://twitter.com/BlocktechCEO/status/1033056549202059264",
            "tweetText": "Dr. Davi Ortega (@ortega_science) introduces the Electron Tomography Database, built by Caltech's @theJensenLab and our team @alexandria, an experiment in collaborative research data sharing built on Open Index Protocol https://www.youtube.com/watch?v=6kzAk79w7PI"
          }
        }
      }
    }
  ],
  "next": "%5B1561877100000%5D"
}
```

#### Get Specific Record
`http://localhost:1606/oip/o5/record/get/fb049475c743a0486129627013d76a4cc83792bce56f4805e8e18d095dd1d5be`

```json
{
  "next": "%5B1561877100000%5D",
  "count": 1,
  "total": 1,
  "results": [
    {
      "meta": {
        "signed_by": "FEQXuxEgfGoEnZbPCHPfHQDMUqXq6tpw4X",
        "publisher_name": "Devon James",
        "block_hash": "ac501599a79a9616b0ad73d441f04a3e21167301552395df6dc1768124fe32df",
        "txid": "fb049475c743a0486129627013d76a4cc83792bce56f4805e8e18d095dd1d5be",
        "block": 3520788,
        "time": 1561877100,
        "type": "oip5"
      },
      "record": {
        "details": {
          "tmpl_90BD561D": {
            "tweetUrl": "https://twitter.com/BlocktechCEO/status/1033056549202059264",
            "tweetText": "Dr. Davi Ortega (@ortega_science) introduces the Electron Tomography Database, built by Caltech's @theJensenLab and our team @alexandria, an experiment in collaborative research data sharing built on Open Index Protocol https://www.youtube.com/watch?v=6kzAk79w7PI"
          }
        }
      }
    }
  ]
}
```

#### Get Template Field Mapping
`http://localhost:1606/oip/o5/record/mapping/tmpl_90BD561D`

```json
{
  "record.details.tmpl_90BD561D.tweetUrl": {
    "full_name": "record.details.tmpl_90BD561D.tweetUrl",
    "mapping": {
      "tweetUrl": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      }
    }
  },
  "record.details.tmpl_90BD561D.ipfsAddressScreenshot.keyword": {
    "mapping": {
      "keyword": {
        "type": "keyword",
        "ignore_above": 256
      }
    },
    "full_name": "record.details.tmpl_90BD561D.ipfsAddressScreenshot.keyword"
  },
  "record.details.tmpl_90BD561D.tweetUrl.keyword": {
    "full_name": "record.details.tmpl_90BD561D.tweetUrl.keyword",
    "mapping": {
      "keyword": {
        "type": "keyword",
        "ignore_above": 256
      }
    }
  },
  "record.details.tmpl_90BD561D.ipfsAddressScreenshot": {
    "full_name": "record.details.tmpl_90BD561D.ipfsAddressScreenshot",
    "mapping": {
      "ipfsAddressScreenshot": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      }
    }
  },
  "record.details.tmpl_90BD561D.tweetText.keyword": {
    "full_name": "record.details.tmpl_90BD561D.tweetText.keyword",
    "mapping": {
      "keyword": {
        "type": "keyword",
        "ignore_above": 256
      }
    }
  },
  "record.details.tmpl_90BD561D.tweetText": {
    "full_name": "record.details.tmpl_90BD561D.tweetText",
    "mapping": {
      "tweetText": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      }
    }
  }
}
```

#### Record Search
`http://localhost:1606/oip/o5/record/search?q=Davi&limit=1`

```json
{
  "count": 1,
  "total": 12,
  "results": [
    {
      "meta": {
        "signed_by": "FEQXuxEgfGoEnZbPCHPfHQDMUqXq6tpw4X",
        "publisher_name": "Devon James",
        "block_hash": "cfdfc3c62f56124ec4b017c38b7d0341e64f931a3833bfb2d871a34a3851a216",
        "txid": "40e961f6cfc207ee30e769ec25aedc4b937063bf454a2afe347078aa12cfd49f",
        "block": 3520957,
        "time": 1561889104,
        "type": "oip5"
      },
      "record": {
        "details": {
          "tmpl_90BD561D": {
            "tweetUrl": "https://twitter.com/BlocktechCEO/status/1033056549202059264",
            "tweetText": "Dr. Davi Ortega (@ortega_science) introduces the Electron Tomography Database, built by Caltech's @theJensenLab and our team @alexandria, an experiment in collaborative research data sharing built on Open Index Protocol https://www.youtube.com/watch?v=6kzAk79w7PI"
          }
        }
      }
    }
  ],
  "next": "%5B1561889104000%2C%2240e961f6cfc207ee30e769ec25aedc4b937063bf454a2afe347078aa12cfd49f%22%5D"
}
```

#### Get Latest Record Templates
`http://localhost:1606/oip/o5/template/get/latest?limit=1&after=%5B1559263656%5D`

```json
{
  "count": 1,
  "total": 26,
  "results": [
    {
      "template": {
        "identifier": 2784722923,
        "friendly_name": "paperclips",
        "file_descriptor_set": "CmcKGG9pcFByb3RvX3RlbXBsYXRlcy5wcm90bxISb2lwUHJvdG8udGVtcGxhdGVzIi8KAVASDQoFY29sb3IYASABKAkSDAoEc2l6ZRgCIAEoCRINCgVicmFuZBgDIAEoCWIGcHJvdG8z",
        "name": "tmpl_A5FB7FEB",
        "description": "styles and brands of paperclips"
      },
      "meta": {
        "signed_by": "FTdQJJCtEP7ZJypXn2RGydebzcFLVgDKXR",
        "block_hash": "ae1ab992c41619825afd39e30a08abc3262c9013391c76231654a756b5543e68",
        "txid": "a5fb7feb6d29af8a40cef438f48980500a305f00735a93a93e87688573d781a0",
        "block": 3449814,
        "time": 1558645926
      }
    }
  ],
  "next": "%5B1558645926%5D"
}
```

#### Get Specific Record Template
`http://localhost:1606/oip/o5/template/get/a5fb7feb`

```json
{
  "total": 1,
  "results": [
    {
      "template": {
        "identifier": 2784722923,
        "friendly_name": "paperclips",
        "file_descriptor_set": "CmcKGG9pcFByb3RvX3RlbXBsYXRlcy5wcm90bxISb2lwUHJvdG8udGVtcGxhdGVzIi8KAVASDQoFY29sb3IYASABKAkSDAoEc2l6ZRgCIAEoCRINCgVicmFuZBgDIAEoCWIGcHJvdG8z",
        "name": "tmpl_A5FB7FEB",
        "description": "styles and brands of paperclips"
      },
      "meta": {
        "signed_by": "FTdQJJCtEP7ZJypXn2RGydebzcFLVgDKXR",
        "block_hash": "ae1ab992c41619825afd39e30a08abc3262c9013391c76231654a756b5543e68",
        "txid": "a5fb7feb6d29af8a40cef438f48980500a305f00735a93a93e87688573d781a0",
        "block": 3449814,
        "time": 1558645926
      }
    }
  ],
  "next": "%5B1558645926%5D",
  "count": 1
}
```

#### Search Record Templates
`http://localhost:1606/oip/o5/template/search?q=paperclips`

```json
{
  "total": 1,
  "results": [
    {
      "template": {
        "identifier": 2784722923,
        "friendly_name": "paperclips",
        "file_descriptor_set": "CmcKGG9pcFByb3RvX3RlbXBsYXRlcy5wcm90bxISb2lwUHJvdG8udGVtcGxhdGVzIi8KAVASDQoFY29sb3IYASABKAkSDAoEc2l6ZRgCIAEoCRINCgVicmFuZBgDIAEoCWIGcHJvdG8z",
        "name": "tmpl_A5FB7FEB",
        "description": "styles and brands of paperclips"
      },
      "meta": {
        "signed_by": "FTdQJJCtEP7ZJypXn2RGydebzcFLVgDKXR",
        "block_hash": "ae1ab992c41619825afd39e30a08abc3262c9013391c76231654a756b5543e68",
        "txid": "a5fb7feb6d29af8a40cef438f48980500a305f00735a93a93e87688573d781a0",
        "block": 3449814,
        "time": 1558645926
      }
    }
  ],
  "next": "%5B1558645926%2C%22a5fb7feb6d29af8a40cef438f48980500a305f00735a93a93e87688573d781a0%22%5D",
  "count": 1
}
```
