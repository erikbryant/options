# Options

Download interesting data about the options for a given ticker.

Save the matching options to files in CSV format and publish them to Google Sheets.

## Build Status

![go fmt](https://github.com/erikbryant/options/actions/workflows/fmt.yml/badge.svg)
![go vet](https://github.com/erikbryant/options/actions/workflows/vet.yml/badge.svg)
![go test](https://github.com/erikbryant/options/actions/workflows/test.yml/badge.svg)

## Glossary

* Ticker - The 1-5 letter symbol for a stock
* Contract - An options contract
* Security - The complete set of data for a given stock, including contracts

## TODO

* Code cleanup
  * More consistent naming. Symbol v ticker v security v option. Ugh!
  * Make it clearer what all the data sources are and where/why each is used.
  * Move more code into `utils`.
    * Code that handles web responses and converts data from interfaces to slices.
  * Separate calculations from printing.
  * Consolidate some of the web request functions.
  * Improve error handling when there are no rows to upload.
    * ```Unable to upload options file: eb_2023-11-17_puts.csv could not create file: invalid argument```
* Add more tests. Split code up so tests are easier to write.
* Add more details to README files.
* Can any of this be put on AWS?
  * Update to latest USE data on weekdays to use up otherwise stranded quota.
  * Generate Google Sheets on weekend; run Fri+Sat, so we get two days of quota.
* Update AES app to also build an encrypt/decrypt app to more easily add API keys.
