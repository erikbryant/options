# Options

Download interesting data about the options for a given ticker.

Print out the options that match the given filters.

Save the matching options to two files in CSV format ready to be loaded into Google Sheets.

## TODO

* Add date of earnings
* Add CSV formulas
* Make throttling pause more intelligently when both providers are throttling
* The cache gets filled with a mix of Finnhub and TradeKing. When next run, Finnhub is
  favored. If there was no entry for Finnhub then it is queried from the web. This is
  slow and wasteful. Figure out how to poll for either being in the cache before going
  to the web. Or, write a backfill that polls Finnhub exclusively so that it never needs
  to flip to TradeKing on a second run. Or, have it check the cache for either F or TK
  before making a call to the web.

* Write a call seller tool
* Have it check upcoming expiration and one subsequent to see if the subsequent expiration
  is more than double the one for this week. If so, add it to the output.

* Write more tests

* Send errors to stderr
* Make it interactive?

* Add list of always-display securities
* Add glossary of terms to README (ticker vs security)

## Glossary

* Ticker - The 1-5 letter symbol for a stock.
* Contract - An options contract.
* Security - The complete set of data for a given stock, including contracts.
