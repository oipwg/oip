API
=

- oipd
  - /version
  - /floData/search?q={query}&limit={limit}
  - /floData/search?q={query}
- artifacts (oip41 & oip042)
  - /artifact/get/latest/{limit:[0-9]+}?nsfw=true/false
  - /artifact/get/latest/{limit:[0-9]+}
  - /artifact/get/{id:[a-f0-9]+}
  - /artifact/search?q={query}&limit={limit}
  - /artifact/search?q={query}
- multipart
  - /multipart/get/ref/{ref:[a-f0-9]+}/{limit:[0-9]+}
  - /multipart/get/ref/{ref:[a-f0-9]+}
  - /multipart/get/id/{id:[a-f0-9]+}
- alexandria-media
  - /alexandria/artifact/get/latest/{limit:[0-9]+}
  - /alexandria/artifact/get/{id:[a-f0-9]+}
- alexandria-publisher
  - /alexandria/publisher/get/latest/{limit:[0-9]+}
  - /alexandria/publisher/get/{address:[A-Za-z0-9]+}
- oip041
  - /oip041/artifact/get/latest/{limit:[0-9]+}?nsfw=true/false
  - /oip041/artifact/get/latest/{limit:[0-9]+}
  - /oip041/artifact/get/{id:[a-f0-9]+}