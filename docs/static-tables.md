## Request Headers

**Slot usage: 101/255**

Complete key-value pairs (Format 1) use a single byte. Name-only headers (Format 2) include the value after the header ID.

| Header ID | Type          | Header Name                    | Header Value                                                                                                                            |
| --------- | ------------- | ------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------- |
| 0x01      | Complete Pair | sec-ch-ua-mobile               | ?0                                                                                                                                      |
| 0x02      | Complete Pair | sec-ch-ua-platform             | "Windows"                                                                                                                               |
| 0x03      | Complete Pair | Accept                         | _/_                                                                                                                                     |
| 0x04      | Complete Pair | accept                         | image/avif,image/webp,image/apng,image/svg+xml,image/_,_/\*;q=0.8                                                                       |
| 0x05      | Complete Pair | X-Requested-With               | XMLHttpRequest                                                                                                                          |
| 0x06      | Complete Pair | Content-Type                   | application/json; charset=UTF-8                                                                                                         |
| 0x07      | Complete Pair | Content-Type                   | text/plain;charset=UTF-8                                                                                                                |
| 0x08      | Complete Pair | sec-ch-ua-arch                 | "x86"                                                                                                                                   |
| 0x09      | Complete Pair | sec-ch-ua-bitness              | "64"                                                                                                                                    |
| 0x0A      | Complete Pair | Sec-GPC                        | 1                                                                                                                                       |
| 0x0B      | Complete Pair | Connection                     | keep-alive                                                                                                                              |
| 0x0C      | Complete Pair | Accept-Language                | en-US,en;q=0.5                                                                                                                          |
| 0x0D      | Complete Pair | Accept-Encoding                | gzip, deflate, br, zstd                                                                                                                 |
| 0x0E      | Complete Pair | Content-Type                   | application/json                                                                                                                        |
| 0x0F      | Complete Pair | Sec-Fetch-Mode                 | cors                                                                                                                                    |
| 0x10      | Complete Pair | Content-Type                   | application/x-www-form-urlencoded                                                                                                       |
| 0x11      | Complete Pair | Sec-Fetch-Site                 | cross-site                                                                                                                              |
| 0x12      | Complete Pair | Sec-Fetch-Site                 | same-origin                                                                                                                             |
| 0x13      | Complete Pair | Sec-Fetch-Dest                 | script                                                                                                                                  |
| 0x14      | Complete Pair | Cache-Control                  | no-cache                                                                                                                                |
| 0x15      | Complete Pair | Pragma                         | no-cache                                                                                                                                |
| 0x16      | Complete Pair | Sec-Fetch-Dest                 | empty                                                                                                                                   |
| 0x17      | Complete Pair | Sec-Fetch-Mode                 | no-cors                                                                                                                                 |
| 0x18      | Complete Pair | cache-control                  | no-cache, no-store                                                                                                                      |
| 0x19      | Complete Pair | Accept                         | text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,_/_;q=0.8,application/signed-exchange;v=b3;q=0.7 |
| 0x1A      | Complete Pair | Upgrade-Insecure-Requests      | 1                                                                                                                                       |
| 0x1B      | Complete Pair | content-type                   | text/plain                                                                                                                              |
| 0x1C      | Complete Pair | Content-Type                   | application/json; charset=utf-8                                                                                                         |
| 0x1D      | Complete Pair | Accept                         | application/json                                                                                                                        |
| 0x1E      | Complete Pair | Sec-Purpose                    | prefetch;prerender                                                                                                                      |
| 0x1F      | Complete Pair | content-type                   | application/json                                                                                                                        |
| 0x20      | Complete Pair | Accept                         | image/avif,image/jxl,image/webp,image/png,image/svg+xml,image/_;q=0.8,_/\*;q=0.5                                                        |
| 0x21      | Complete Pair | Sec-Fetch-Dest                 | image                                                                                                                                   |
| 0x22      | Complete Pair | Sec-Fetch-Site                 | same-site                                                                                                                               |
| 0x23      | Complete Pair | Sec-Purpose                    | prefetch                                                                                                                                |
| 0x24      | Complete Pair | accept                         | image/webp,_/_                                                                                                                          |
| 0x25      | Complete Pair | Accept                         | application/json, text/plain, _/_                                                                                                       |
| 0x26      | Complete Pair | Accept                         | application/json, text/javascript, _/_; q=0.01                                                                                          |
| 0x27      | Complete Pair | Content-Encoding               | gzip                                                                                                                                    |
| 0x28      | Complete Pair | Service-Worker                 | script                                                                                                                                  |
| 0x29      | Complete Pair | Content-Type                   | application/x-www-form-urlencoded; charset=UTF-8                                                                                        |
| 0x2A      | Complete Pair | Content-Type                   | application/json+protobuf                                                                                                               |
| 0x2B      | Complete Pair | Sec-Fetch-Mode                 | same-origin                                                                                                                             |
| 0x2C      | Complete Pair | Access-Control-Request-Method  | POST                                                                                                                                    |
| 0x2D      | Complete Pair | Sec-Fetch-Dest                 | style                                                                                                                                   |
| 0x2E      | Complete Pair | Accept                         | application/signed-exchange;v=b3;q=0.7,_/_;q=0.8                                                                                        |
| 0x2F      | Complete Pair | Accept                         | text/html                                                                                                                               |
| 0x30      | Complete Pair | Accept                         | application/font-woff2;q=1.0,application/font-woff;q=0.9,_/_;q=0.8                                                                      |
| 0x31      | Complete Pair | Sec-Fetch-Dest                 | font                                                                                                                                    |
| 0x32      | Complete Pair | accept                         | text/event-stream                                                                                                                       |
| 0x33      | Complete Pair | Sec-Fetch-Mode                 | navigate                                                                                                                                |
| 0x34      | Complete Pair | Content-Length                 | 0                                                                                                                                       |
| 0x35      | Complete Pair | Connection                     | Upgrade                                                                                                                                 |
| 0x36      | Complete Pair | Upgrade                        | websocket                                                                                                                               |
| 0x37      | Complete Pair | Cache-Control                  | max-age=0                                                                                                                               |
| 0x38      | Complete Pair | Range                          | bytes=0-                                                                                                                                |
| 0x39      | Complete Pair | Sec-Fetch-Dest                 | document                                                                                                                                |
| 0x3A      | Complete Pair | Sec-Fetch-Mode                 | websocket                                                                                                                               |
| 0x3B      | Complete Pair | Sec-Fetch-User                 | ?1                                                                                                                                      |
| 0x3C      | Complete Pair | Access-Control-Request-Method  | GET                                                                                                                                     |
| 0x3D      | Complete Pair | sec-ch-ua-mobile               | ?1                                                                                                                                      |
| 0x3E      | Complete Pair | sec-ch-ua-platform             | "macOS"                                                                                                                                 |
| 0x3F      | Complete Pair | sec-ch-ua-platform             | "Linux"                                                                                                                                 |
| 0x40      | Complete Pair | sec-ch-ua-platform             | "Android"                                                                                                                               |
| 0x41      | Complete Pair | sec-ch-ua-arch                 | "arm"                                                                                                                                   |
| 0x42      | Name Only     | User-Agent                     | (variable)                                                                                                                              |
| 0x43      | Name Only     | sec-ch-ua-mobile               | (variable)                                                                                                                              |
| 0x44      | Name Only     | sec-ch-ua-platform             | (variable)                                                                                                                              |
| 0x45      | Name Only     | sec-ch-ua                      | (variable)                                                                                                                              |
| 0x46      | Name Only     | Accept                         | (variable)                                                                                                                              |
| 0x47      | Name Only     | Content-Type                   | (variable)                                                                                                                              |
| 0x48      | Name Only     | accept                         | (variable)                                                                                                                              |
| 0x49      | Name Only     | sec-ch-ua-platform-version     | (variable)                                                                                                                              |
| 0x4A      | Name Only     | sec-ch-ua-arch                 | (variable)                                                                                                                              |
| 0x4B      | Name Only     | Connection                     | (variable)                                                                                                                              |
| 0x4C      | Name Only     | Host                           | (variable)                                                                                                                              |
| 0x4D      | Name Only     | Sec-GPC                        | (variable)                                                                                                                              |
| 0x4E      | Name Only     | Accept-Language                | (variable)                                                                                                                              |
| 0x4F      | Name Only     | Sec-Fetch-Mode                 | (variable)                                                                                                                              |
| 0x50      | Name Only     | Sec-Fetch-Site                 | (variable)                                                                                                                              |
| 0x51      | Name Only     | Sec-Fetch-Dest                 | (variable)                                                                                                                              |
| 0x52      | Name Only     | Accept-Encoding                | (variable)                                                                                                                              |
| 0x53      | Name Only     | Referer                        | (variable)                                                                                                                              |
| 0x54      | Name Only     | Authorization                  | (variable)                                                                                                                              |
| 0x55      | Name Only     | Origin                         | (variable)                                                                                                                              |
| 0x56      | Name Only     | Cookie                         | (variable)                                                                                                                              |
| 0x57      | Name Only     | Cache-Control                  | (variable)                                                                                                                              |
| 0x58      | Name Only     | Pragma                         | (variable)                                                                                                                              |
| 0x59      | Name Only     | Sec-Purpose                    | (variable)                                                                                                                              |
| 0x5A      | Name Only     | Content-Length                 | (variable)                                                                                                                              |
| 0x5B      | Name Only     | prefer                         | (variable)                                                                                                                              |
| 0x5C      | Name Only     | Content-Encoding               | (variable)                                                                                                                              |
| 0x5D      | Name Only     | Access-Control-Request-Method  | (variable)                                                                                                                              |
| 0x5E      | Name Only     | Access-Control-Request-Headers | (variable)                                                                                                                              |
| 0x5F      | Name Only     | Range                          | (variable)                                                                                                                              |
| 0x60      | Name Only     | Sec-WebSocket-Version          | (variable)                                                                                                                              |
| 0x61      | Name Only     | Upgrade                        | (variable)                                                                                                                              |
| 0x62      | Name Only     | If-None-Match                  | (variable)                                                                                                                              |
| 0x63      | Name Only     | Sec-WebSocket-Extensions       | (variable)                                                                                                                              |
| 0x64      | Name Only     | Sec-WebSocket-Key              | (variable)                                                                                                                              |
| 0x65      | Name Only     | If-Modified-Since              | (variable)                                                                                                                              |

## Response Headers

**Slot usage: 191/255**

Complete key-value pairs (Format 1) use a single byte. Name-only headers (Format 2) include the value after the header ID.

| Header ID | Type          | Header Name                         | Header Value                                            |
| --------- | ------------- | ----------------------------------- | ------------------------------------------------------- |
| 0x01      | Complete Pair | x-content-type-options              | nosniff                                                 |
| 0x02      | Complete Pair | timing-allow-origin                 | \*                                                      |
| 0x03      | Complete Pair | access-control-allow-origin         | \*                                                      |
| 0x04      | Complete Pair | vary                                | Accept-Encoding                                         |
| 0x05      | Complete Pair | content-encoding                    | gzip                                                    |
| 0x06      | Complete Pair | strict-transport-security           | max-age=31536000                                        |
| 0x07      | Complete Pair | cross-origin-resource-policy        | cross-origin                                            |
| 0x08      | Complete Pair | content-encoding                    | br                                                      |
| 0x09      | Complete Pair | x-download-options                  | noopen                                                  |
| 0x0A      | Complete Pair | pragma                              | no-cache                                                |
| 0x0B      | Complete Pair | accept-ranges                       | bytes                                                   |
| 0x0C      | Complete Pair | content-type                        | application/json; charset=utf-8                         |
| 0x0D      | Complete Pair | content-disposition                 | attachment                                              |
| 0x0E      | Complete Pair | x-xss-protection                    | 0                                                       |
| 0x0F      | Complete Pair | cache-control                       | no-cache                                                |
| 0x10      | Complete Pair | cache-control                       | private                                                 |
| 0x11      | Complete Pair | content-length                      | 0                                                       |
| 0x12      | Complete Pair | content-type                        | application/javascript                                  |
| 0x13      | Complete Pair | x-frame-options                     | SAMEORIGIN                                              |
| 0x14      | Complete Pair | content-type                        | image/gif                                               |
| 0x15      | Complete Pair | strict-transport-security           | max-age=31536000; includeSubDomains                     |
| 0x16      | Complete Pair | content-type                        | image/avif                                              |
| 0x17      | Complete Pair | content-type                        | application/json                                        |
| 0x18      | Complete Pair | strict-transport-security           | max-age=31536000; includeSubDomains; preload            |
| 0x19      | Complete Pair | vary                                | Origin                                                  |
| 0x1A      | Complete Pair | x-xss-protection                    | 1; mode=block                                           |
| 0x1B      | Complete Pair | cf-cache-status                     | HIT                                                     |
| 0x1C      | Complete Pair | cache-control                       | max-age=630720000                                       |
| 0x1D      | Complete Pair | cache-control                       | max-age=86400000                                        |
| 0x1E      | Complete Pair | referrer-policy                     | strict-origin-when-cross-origin                         |
| 0x1F      | Complete Pair | cf-cache-status                     | DYNAMIC                                                 |
| 0x20      | Complete Pair | strict-transport-security           | max-age=0                                               |
| 0x21      | Complete Pair | cache-control                       | public, max-age=31536000                                |
| 0x22      | Complete Pair | access-control-allow-methods        | GET, POST, OPTIONS                                      |
| 0x23      | Complete Pair | expires                             | Thu, 01 Jan 1970 00:00:01 GMT                           |
| 0x24      | Complete Pair | content-type                        | text/plain                                              |
| 0x25      | Complete Pair | expires                             | -1                                                      |
| 0x26      | Complete Pair | vary                                | accept-encoding                                         |
| 0x27      | Complete Pair | cross-origin-opener-policy          | same-origin-allow-popups                                |
| 0x28      | Complete Pair | content-type                        | text/javascript                                         |
| 0x29      | Complete Pair | content-type                        | text/html; charset=UTF-8                                |
| 0x2A      | Complete Pair | expires                             | Fri, 01 Jan 1990 00:00:00 GMT                           |
| 0x2B      | Complete Pair | x-permitted-cross-domain-policies   | none                                                    |
| 0x2C      | Complete Pair | Content-Type                        | text/css                                                |
| 0x2D      | Complete Pair | permissions-policy                  | unload=()                                               |
| 0x2E      | Complete Pair | content-type                        | image/jpeg                                              |
| 0x2F      | Complete Pair | cache-control                       | no-cache, no-store, must-revalidate                     |
| 0x30      | Complete Pair | vary                                | Accept-Encoding, Origin                                 |
| 0x31      | Complete Pair | cache-control                       | no-store, no-cache                                      |
| 0x32      | Complete Pair | cache-control                       | public, max-age=2592000                                 |
| 0x33      | Complete Pair | referrer-policy                     | no-referrer-when-downgrade                              |
| 0x34      | Complete Pair | x-frame-options                     | DENY                                                    |
| 0x35      | Complete Pair | access-control-allow-methods        | POST                                                    |
| 0x36      | Complete Pair | content-type                        | text/javascript; charset=UTF-8                          |
| 0x37      | Complete Pair | strict-transport-security           | max-age=63072000                                        |
| 0x38      | Complete Pair | cache-control                       | public, immutable, max-age=31536000                     |
| 0x39      | Complete Pair | Connection                          | keep-alive                                              |
| 0x3A      | Complete Pair | content-type                        | image/webp                                              |
| 0x3B      | Complete Pair | cache-control                       | public,max-age=31536000                                 |
| 0x3C      | Complete Pair | Cache-Control                       | public, max-age=31536000, immutable                     |
| 0x3D      | Complete Pair | content-type                        | image/png                                               |
| 0x3E      | Complete Pair | cache-control                       | max-age=0, private, must-revalidate                     |
| 0x3F      | Complete Pair | permissions-policy                  | interest-cohort=()                                      |
| 0x40      | Complete Pair | expires                             | 0                                                       |
| 0x41      | Complete Pair | cache-control                       | no-cache, must-revalidate                               |
| 0x42      | Complete Pair | strict-transport-security           | max-age=63072000; includeSubDomains; preload            |
| 0x43      | Complete Pair | content-encoding                    | zstd                                                    |
| 0x44      | Complete Pair | access-control-expose-headers       | \*                                                      |
| 0x45      | Complete Pair | content-type                        | application/javascript; charset=utf-8                   |
| 0x46      | Complete Pair | cache-control                       | no-cache, no-store                                      |
| 0x47      | Complete Pair | access-control-max-age              | 7200                                                    |
| 0x48      | Complete Pair | vary                                | Accept, Origin                                          |
| 0x49      | Complete Pair | access-control-allow-methods        | GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS            |
| 0x4A      | Complete Pair | cache-control                       | max-age=2592000,s-maxage=86400                          |
| 0x4B      | Complete Pair | content-type                        | image/svg+xml                                           |
| 0x4C      | Complete Pair | access-control-max-age              | 86400                                                   |
| 0x4D      | Complete Pair | content-disposition                 | attachment; filename="f.txt"                            |
| 0x4E      | Complete Pair | x-robots-tag                        | none                                                    |
| 0x4F      | Complete Pair | cache-control                       | public, max-age=22222222                                |
| 0x50      | Complete Pair | cross-origin-opener-policy          | unsafe-none                                             |
| 0x51      | Complete Pair | strict-transport-security           | max-age=15552000; includeSubDomains; preload            |
| 0x52      | Complete Pair | content-type                        | text/css; charset=UTF-8                                 |
| 0x53      | Complete Pair | x-frame-options                     | deny                                                    |
| 0x54      | Complete Pair | Content-Type                        | text/html                                               |
| 0x55      | Complete Pair | access-control-allow-methods        | OPTIONS,GET,POST                                        |
| 0x56      | Complete Pair | cache-control                       | public,max-age=31536000,immutable                       |
| 0x57      | Complete Pair | cache-control                       | public, max-age=31536000, immutable                     |
| 0x58      | Complete Pair | access-control-allow-methods        | GET, HEAD                                               |
| 0x59      | Complete Pair | cache-control                       | max-age=31536000                                        |
| 0x5A      | Complete Pair | x-dns-prefetch-control              | on                                                      |
| 0x5B      | Complete Pair | content-type                        | application/json;charset=UTF-8                          |
| 0x5C      | Complete Pair | cache-control                       | private, max-age=3600                                   |
| 0x5D      | Complete Pair | access-control-allow-methods        | GET                                                     |
| 0x5E      | Complete Pair | content-type                        | font/woff2                                              |
| 0x5F      | Complete Pair | access-control-max-age              | 3000                                                    |
| 0x60      | Complete Pair | Cache-Control                       | public                                                  |
| 0x61      | Complete Pair | Cache-Control                       | max-age=31536000, public                                |
| 0x62      | Complete Pair | permissions-policy                  | microphone=()                                           |
| 0x63      | Complete Pair | Transfer-Encoding                   | chunked                                                 |
| 0x64      | Complete Pair | cache-control                       | no-cache, no-store, max-age=0, must-revalidate          |
| 0x65      | Complete Pair | cache-control                       | max-age=300                                             |
| 0x66      | Complete Pair | cache-control                       | max-age=2592000                                         |
| 0x67      | Complete Pair | cache-control                       | no-store                                                |
| 0x68      | Complete Pair | cf-cache-status                     | MISS                                                    |
| 0x69      | Complete Pair | cache-control                       | public, max-age=7200                                    |
| 0x6A      | Complete Pair | cache-control                       | public,max-age=604800                                   |
| 0x6B      | Complete Pair | content-type                        | text/plain; charset=utf-8                               |
| 0x6C      | Complete Pair | cache-control                       | private, no-cache, no-store, max-age=0, must-revalidate |
| 0x6D      | Complete Pair | cross-origin-embedder-policy        | require-corp                                            |
| 0x6E      | Complete Pair | cache-control                       | private, max-age=0                                      |
| 0x6F      | Complete Pair | origin-agent-cluster                | ?1                                                      |
| 0x70      | Complete Pair | vary                                | Accept, Accept-Encoding                                 |
| 0x71      | Complete Pair | age                                 | 0                                                       |
| 0x72      | Complete Pair | cache-control                       | public, max-age=86400                                   |
| 0x73      | Complete Pair | cache-control                       | max-age=0, no-cache, no-store                           |
| 0x74      | Complete Pair | pragma                              | public                                                  |
| 0x75      | Complete Pair | accept-ranges                       | none                                                    |
| 0x76      | Complete Pair | cache-control                       | max-age=2147483648, immutable                           |
| 0x77      | Complete Pair | vary                                | Referer                                                 |
| 0x78      | Complete Pair | cache-control                       | public, max-age=604800                                  |
| 0x79      | Complete Pair | cache-control                       | public, max-age=86400, no-transform                     |
| 0x7A      | Complete Pair | cache-control                       | public, max-age=0, must-revalidate                      |
| 0x7B      | Complete Pair | vary                                | Origin, Accept-Encoding                                 |
| 0x7C      | Complete Pair | cache-control                       | no-store, no-cache, must-revalidate                     |
| 0x7D      | Complete Pair | vary                                | Accept                                                  |
| 0x7E      | Complete Pair | cache-control                       | private, max-age=900                                    |
| 0x7F      | Complete Pair | content-security-policy             | frame-ancestors 'self'                                  |
| 0x80      | Complete Pair | content-type                        | application/json+protobuf; charset=UTF-8                |
| 0x81      | Complete Pair | cross-origin-opener-policy          | same-origin                                             |
| 0x82      | Complete Pair | cross-origin-resource-policy        | same-origin                                             |
| 0x83      | Complete Pair | Content-Type                        | application/octet-stream                                |
| 0x84      | Complete Pair | content-type                        | application/json; odata.metadata=minimal                |
| 0x85      | Complete Pair | content-type                        | text/plain;charset=UTF-8                                |
| 0x86      | Complete Pair | content-type                        | video/MP2T                                              |
| 0x87      | Complete Pair | content-type                        | application/javascript;charset=utf-8                    |
| 0x88      | Complete Pair | cross-origin-resource-policy        | same-site                                               |
| 0x89      | Complete Pair | content-type                        | video/x-m4v                                             |
| 0x8A      | Complete Pair | Upgrade                             | websocket                                               |
| 0x8B      | Complete Pair | Content-Type                        | text/xml                                                |
| 0x8C      | Complete Pair | content-type                        | application/font-woff                                   |
| 0x8D      | Complete Pair | content-type                        | audio/mpeg                                              |
| 0x8E      | Complete Pair | content-type                        | text/javascript;charset=UTF-8                           |
| 0x8F      | Name Only     | date                                | (variable)                                              |
| 0x90      | Name Only     | content-type                        | (variable)                                              |
| 0x91      | Name Only     | cache-control                       | (variable)                                              |
| 0x92      | Name Only     | content-length                      | (variable)                                              |
| 0x93      | Name Only     | server                              | (variable)                                              |
| 0x94      | Name Only     | access-control-allow-origin         | (variable)                                              |
| 0x95      | Name Only     | strict-transport-security           | (variable)                                              |
| 0x96      | Name Only     | vary                                | (variable)                                              |
| 0x97      | Name Only     | content-encoding                    | (variable)                                              |
| 0x98      | Name Only     | x-content-type-options              | (variable)                                              |
| 0x99      | Name Only     | timing-allow-origin                 | (variable)                                              |
| 0x9A      | Name Only     | report-to                           | (variable)                                              |
| 0x9B      | Name Only     | alt-svc                             | (variable)                                              |
| 0x9C      | Name Only     | last-modified                       | (variable)                                              |
| 0x9D      | Name Only     | access-control-expose-headers       | (variable)                                              |
| 0x9E      | Name Only     | expires                             | (variable)                                              |
| 0x9F      | Name Only     | age                                 | (variable)                                              |
| 0xA0      | Name Only     | cross-origin-resource-policy        | (variable)                                              |
| 0xA1      | Name Only     | x-xss-protection                    | (variable)                                              |
| 0xA2      | Name Only     | etag                                | (variable)                                              |
| 0xA3      | Name Only     | content-disposition                 | (variable)                                              |
| 0xA4      | Name Only     | pragma                              | (variable)                                              |
| 0xA5      | Name Only     | via                                 | (variable)                                              |
| 0xA6      | Name Only     | accept-ranges                       | (variable)                                              |
| 0xA7      | Name Only     | x-download-options                  | (variable)                                              |
| 0xA8      | Name Only     | x-frame-options                     | (variable)                                              |
| 0xA9      | Name Only     | accept-ch                           | (variable)                                              |
| 0xAA      | Name Only     | access-control-allow-credentials    | (variable)                                              |
| 0xAB      | Name Only     | access-control-allow-methods        | (variable)                                              |
| 0xAC      | Name Only     | content-security-policy             | (variable)                                              |
| 0xAD      | Name Only     | referrer-policy                     | (variable)                                              |
| 0xAE      | Name Only     | access-control-allow-headers        | (variable)                                              |
| 0xAF      | Name Only     | permissions-policy                  | (variable)                                              |
| 0xB0      | Name Only     | content-security-policy-report-only | (variable)                                              |
| 0xB1      | Name Only     | access-control-max-age              | (variable)                                              |
| 0xB2      | Name Only     | x-permitted-cross-domain-policies   | (variable)                                              |
| 0xB3      | Name Only     | Connection                          | (variable)                                              |
| 0xB4      | Name Only     | x-robots-tag                        | (variable)                                              |
| 0xB5      | Name Only     | location                            | (variable)                                              |
| 0xB6      | Name Only     | link                                | (variable)                                              |
| 0xB7      | Name Only     | set-cookie                          | (variable)                                              |
| 0xB8      | Name Only     | origin-agent-cluster                | (variable)                                              |
| 0xB9      | Name Only     | content-language                    | (variable)                                              |
| 0xBA      | Name Only     | cross-origin-embedder-policy        | (variable)                                              |
| 0xBB      | Name Only     | www-authenticate                    | (variable)                                              |
| 0xBC      | Name Only     | Content-Range                       | (variable)                                              |
| 0xBD      | Name Only     | retry-after                         | (variable)                                              |
| 0xBE      | Name Only     | critical-ch                         | (variable)                                              |
| 0xBF      | Name Only     | X-Payment-Response                  | (variable)                                              |
