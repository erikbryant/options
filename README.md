# Options

Download interesting data about the options for a given ticker.

## Uses the Yahoo! Finance backend

Get option data for a single ticker (Ford): https://finance.yahoo.com/quote/F/options?p=F

Get option data for a single strike on a single expiration date (F @ 6.5): https://finance.yahoo.com/quote/F/options?strike=6.5&straddle=false&date=1597363200

## TODO

Write tests
* yahoo
  * get()
  * ParseContracts()
  * ParsePrice()
  * ParseStrikes()

* Send errors to stderr
* Add sorting flags
* Adjust bid/strike ratio to account for time-to-expiration
* Take ticker list input from file
* Make it interactive?
