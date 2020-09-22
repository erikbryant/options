# The Yahoo! Finance backend

Get option data for a single ticker (Ford): https://finance.yahoo.com/quote/F/options?p=F

Get option data for a single strike on a single expiration date (F @ 6.5): https://finance.yahoo.com/quote/F/options?strike=6.5&straddle=false&date=1597363200

It turns out that Yahoo's data is very dirty and flaky. Sometimes it will return an empty set even though data does exist. This condition is very hard to test for since there is no associated error and an empty set is a legal value. Because of this and other data errors I have disabled this as a backend. Maybe someday we can come back to it.
