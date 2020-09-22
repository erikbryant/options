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
  to flip to TradeKing on a second run.
* If we run this on a Friday evening we still get data output for that day's expirations,
  which are no longer available. Either check the time of day to see if the market is
  still open or make the expiration flag a date range.

* Write a call seller tool
* Have it check upcoming expiration and one subsequent to see if the subsequent expiration
  is more than double the one for this week. If so, add it to the output.

* Write more tests

* Send errors to stderr
* Make it interactive?
