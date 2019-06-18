API
=

- oipd
  - /version
  - /floData/search?q={query}
- artifacts (oip41 & oip042)
  - /artifact/get/latest?nsfw=true/false
  - /artifact/get/{id:[a-f0-9]+}
  - /artifact/search?q={query}
- multipart
  - /multipart/get/ref/{ref:[a-f0-9]+}
  - /multipart/get/id/{id:[a-f0-9]+}
- alexandria-media
  - /alexandria/artifact/get/latest
  - /alexandria/artifact/get/{id:[a-f0-9]+}
- alexandria-publisher
  - /alexandria/publisher/get/latest
  - /alexandria/publisher/get/{address:[A-Za-z0-9]+}
- oip041
  - /oip041/artifact/get/latest?nsfw=true/false
  - /oip041/artifact/get/{id:[a-f0-9]+}
- oip042
  - /oip042/artifact/get/latest?nsfw=true/false
  - /oip042/record/get/{originalTxid}
  - /oip042/record/get/{originalTxid}/version/{editRecordTxid}
  - /oip042/edit/get/{editRecordTxid}
  - /oip042/edit/search?q={query}


##Common Query Params
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
