# marketData options API

Dashboard, including quota used: [Dashboard](https://www.marketdata.app/dashboard/)

## Request stats

582 tickers @ strikeLimit=4 + 10 tickers that 404 = 4640 requests

## TODO

* Reduce number of requests
  * Dynamically set the number of strikes to get per each symbol.
  * Eliminate more symbols?
  * Use THETA or BID to know when to stop requesting strikes.
* In marketData.GetOptions(), instead of a single, specific expiration, get all expirations up to and including that passed-in expiration date.
* Fix the logic that determines which symbols have weekly (or less) expirations. Run it and update options.csv
* Uncomment the `regenerate` line that deletes old entries from the web cache.
* Cache improvements
  * Make cahcing smarter about when to update or not.
  * We have data cached from finnhub, but can't use it because we think we are getting from TradeKing. Maybe save the internal data instead of the raw data? Maybe check the cache for either format before checking the web?
* Code cleanup
  * More consistent naming. Symbol v ticker v security v option. Ugh!
  * Make it clearer what all the data sources are and where/why each is used.
  * Move more code into `utils`.
    * Code that handles web responses and converts data from interfaces to slices.
  * Separate calcuations from printing.
  * Separate printing from posting to gDrive.
  * Consolidate some of the web request functions.
  * Improve error handling when there are no rows to upload.
    * ```Unable to upload options file: eb_2023-11-17_puts.csv could not create file: invalid argument```
* Add more tests. Split code up so tests are easier to write.
* Add more details to README files.
* Can any of this be put on AWS?
  * Update to latest USE data on weekdays to use up otherwise stranded quota.
  * Generate Google Sheets on weekend; run Fri+Sat so we get two days of quota.
* Update AES app to also build an encrypt/decrypt app to more easily add API keys.
* Fix the 404 error handling. (see below)

```text
AMRS    Error getting security data error getting options AMRS error concatenating marketData options AMRS error concatenating marketData option data got an unexpected StatusCode &{404 Not Found 404 HTTP/2.0 2 0 map[Allow:[GET, HEAD, OPTIONS] Alt-Svc:[h3=":443"; ma=86400] Cf-Cache-Status:[DYNAMIC] Cf-Ray:[8278bce81b9db15d-ATL] Content-Length:[43] Content-Type:[application/json] Cross-Origin-Opener-Policy:[same-origin] Date:[Fri, 17 Nov 2023 14:42:02 GMT] Nel:[{"success_fraction":0,"report_to":"cf-nel","max_age":604800}] Referrer-Policy:[same-origin] Report-To:[{"endpoints":[{"url":"https:\/\/a.nel.cloudflare.com\/report\/v3?s=82Y3dfD8pJtCAmU736NWPnAuk4FOKQ7%2FHbuJbYcVvol5ltlklH8jBFg0pRf61SfHXil175OFp82%2BBEog8o02B2B%2F7pgI3phh6j%2BAgIQ0KFPNlSAFu76ZI4fd9BS%2FQpA6qcsvtEaZSGEHrOKqWCFwbnk%3D"}],"group":"cf-nel","max_age":604800}] Server:[cloudflare] Vary:[Accept, Origin] X-Api-Ratelimit-Consumed:[0] X-Api-Ratelimit-Limit:[10000] X-Api-Ratelimit-Remaining:[9784] X-Api-Ratelimit-Reset:[1700317800] X-Api-Response-Log-Id:[52693856] X-Content-Type-Options:[nosniff] X-Frame-Options:[DENY]] {0x140003ec180} 43 [] false false map[] 0x14000188300 0x140003ea0b0}
```
