# marketData options API

[Dashboard](https://www.marketdata.app/dashboard/), including quota used

[marketData API](https://docs.marketdata.app/api)

## TODO

* Reduce number of requests
  * Dynamically set the number of strikes to get per each symbol.
  * Eliminate more symbols?
  * Use THETA or BID to know when to stop requesting strikes.
* Code cleanup
  * More consistent naming. Symbol v ticker v security v option. Ugh!
  * Make it clearer what all the data sources are and where/why each is used.
  * Move more code into `utils`.
    * Code that handles web responses and converts data from interfaces to slices.
  * Separate calcuations from printing.
  * Consolidate some of the web request functions.
  * Improve error handling when there are no rows to upload.
    * ```Unable to upload options file: eb_2023-11-17_puts.csv could not create file: invalid argument```
* Add more tests. Split code up so tests are easier to write.
* Add more details to README files.
* Can any of this be put on AWS?
  * Update to latest USE data on weekdays to use up otherwise stranded quota.
  * Generate Google Sheets on weekend; run Fri+Sat so we get two days of quota.
* Update AES app to also build an encrypt/decrypt app to more easily add API keys.
