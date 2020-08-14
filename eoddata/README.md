# eoddata

https://www.eoddata.com/ provides downloadable lists of all US equities. While it is possible to download the files programmatically, it is not recommended. There is a cookie needed and the cookie rotates too often.

## Request URL

The download URL is of the form

```
https://www.eoddata.com/data/filedownload.aspx?g=1&sd=20200813&ed=20200813&d=4&p=0&o=d&k=vubvanxsz4
```

Where

* `g` the exchange group (1=US equities)
* `sd` and `ed` are the start and end data of the range
* `d` is the file format to use (4=CSV)
* `p` period ???
* `o` and `k` are the download key

## Download key

The URL contains a download key of the form `o=d&k=vubvanxsz4`. This key rotates every few minutes. Without this key, a download request will receive a zero-sized response.

## Headers

When making a download request the `cookie` header must be provided. No other headers are needed.
