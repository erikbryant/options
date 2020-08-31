# Options

Download interesting data about the options for a given ticker.

## Uses the Yahoo! Finance backend

Get option data for a single ticker (Ford): https://finance.yahoo.com/quote/F/options?p=F

Get option data for a single strike on a single expiration date (F @ 6.5): https://finance.yahoo.com/quote/F/options?strike=6.5&straddle=false&date=1597363200

## TODO

* Write tests

* Use Finnhub for current share price (flip between it and TradeKing?)
* Add flag to invalidate cache for selected tickers
* If fewer than XX tickers, delete cache and get live data

* Add columns for quantity, exposure, yield
* Add columns for exposure?

* Send errors to stderr
* Make it interactive?
