package finnhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/options/cache"
	"github.com/erikbryant/options/date"
	"github.com/erikbryant/options/security"
	"github.com/erikbryant/web"
)

var (
	cipherAuthToken = "GDrwdFOt/zTS12HUuCYE82Xjdzoa5EYOT9e377XNYc0w2St/CNQ0M/jOZorYzU1G"
	authToken       = ""
	earnings        map[string]string
)

// Init initializes the internal state of the package
func Init(passPhrase, latestExpiration string) {
	var err error

	authToken, err = aes.Decrypt(cipherAuthToken, passPhrase)
	if err != nil {
		panic("Incorrect passphrase for FinnHub")
	}

	earnings, err = earningDates(latestExpiration)
	if err != nil {
		panic("Unable to get earnings dates")
	}
}

func Earnings(symbol string) string {
	return earnings[symbol]
}

func webRequest(url string) (map[string]interface{}, bool, error) {
	var response *http.Response
	var err error

	// API key authentication
	url += "&token=" + authToken

	for {
		response, err = web.Request2(url, map[string]string{})
		if err != nil {
			return nil, false, fmt.Errorf("error fetching symbol data %s", err)
		}
		if response.StatusCode == 429 {
			reset, err := strconv.ParseInt(response.Header["X-Ratelimit-Reset"][0], 10, 64)
			if err != nil {
				return nil, true, fmt.Errorf("throttled")
			}
			seconds := reset - time.Now().Unix()
			if seconds <= 0 || seconds > 10 {
				return nil, true, fmt.Errorf("throttled")
			}
			after, err := time.ParseDuration(fmt.Sprintf("%ds", seconds))
			if err != nil || after < 5 {
				after = time.Duration(5 * time.Second)
			}
			fmt.Printf("Throttled. Backing off for %s...", after)
			time.Sleep(after)
			fmt.Printf("done\n")
			continue
		}
		if response.StatusCode == 200 {
			break
		}
		return nil, false, fmt.Errorf("got an unexpected StatusCode %v", response)
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, false, err
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return nil, false, fmt.Errorf("unable to unmarshal json %s", err)
	}

	return jsonObject, false, nil
}

// parseEarnings parses the earnings json returned from finnhub
func parseEarnings(m map[string]interface{}) (map[string]string, error) {
	earnings := make(map[string]string)

	earningsCalendar, ok := m["earningsCalendar"]
	if !ok {
		return nil, fmt.Errorf("unable to parse earningsCalendar object")
	}

	for _, val := range earningsCalendar.([]interface{}) {
		symbol, ok := val.(map[string]interface{})["symbol"]
		if !ok {
			return nil, fmt.Errorf("unable to parse symbol object")
		}

		d, ok := val.(map[string]interface{})["date"]
		if !ok {
			return nil, fmt.Errorf("unable to parse date object")
		}

		earnings[symbol.(string)] = d.(string)
	}

	return earnings, nil
}

// parseQuote parses the quote json returned from finnhub
func parseQuote(m map[string]interface{}, sec *security.Security) error {
	// {
	//   "c": 207.39,      close
	//   "d": -3.99,       delta (close - previousClose)
	//   "dp": -1.8876,    delta percent (delta / previousClose)
	//   "h": 227.3,       high
	//   "l": 205.6,       low
	//   "o": 213.41,      open
	//   "pc": 211.38,     previousClose
	//   "t": 1709931601   timestamp
	// }
	var ok bool

	t, err := web.MsiValue(m, []string{"t"})
	if err != nil {
		return fmt.Errorf("unable to parse quote object timestamp %s", err)
	}

	c, err := web.MsiValue(m, []string{"c"})
	if err != nil {
		return fmt.Errorf("unable to parse quote object close %s", err)
	}

	sec.Price, ok = c.(float64)
	if !ok {
		return fmt.Errorf("unable to convert c to float64 %v", c)
	}

	now := time.Now()
	quoteDate := time.Unix(int64(t.(float64)), 0)
	sinceClose := date.TimeSinceClose(now)
	stale := now.Sub(quoteDate) > (sinceClose + (24+6)*time.Hour + 30*time.Minute)

	if c.(float64) == 0 || t.(float64) == 0 || stale {
		msg := "From Finnhub"
		if c.(float64) == 0 {
			msg += ", sec price is zero"
		}
		if t.(float64) == 0 {
			msg += ", sec timestamp is zero"
		}
		if stale {
			msg += ", sec timestamp is stale"
		}
		fmt.Println(msg)
		sec.Price = 0.0
	}

	return nil
}

// parseMetric parses the quote json returned from finnhub
func parseMetric(m map[string]interface{}, sec *security.Security) error {
	metric, err := web.MsiValue(m, []string{"metric"})
	if err != nil {
		return fmt.Errorf("unable to parse metric object %s", err)
	}

	// The pe key is optional; ignore it if not there
	// When present, sometimes it is nil; remap that to zero
	pe, err := web.MsiValued(metric, []string{"peBasicExclExtraTTM"}, 0.0)
	if err != nil {
		return nil
	}

	sec.PE = pe.(float64)

	return nil
}

func fetch(url string) (map[string]interface{}, error) {
	response, err := cache.Read(url)
	if err == nil {
		return response, nil
	}

	retry := true
	for retry {
		response, retry, err = webRequest(url)
	}
	if err != nil {
		return nil, err
	}

	cache.Update(url, response)

	return response, nil
}

// earningDates returns all earning announcement dates from now to the end date
func earningDates(end string) (map[string]string, error) {
	start := time.Now().Format("2006-01-02")

	url := "https://finnhub.io/api/v1/calendar/earnings?from=" + start + "&to=" + end

	response, err := fetch(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching earnings dates %s", err)
	}

	dates, err := parseEarnings(response)
	if err != nil {
		return nil, fmt.Errorf("error parsing earning dates %s", err)
	}

	return dates, nil
}

func getQuote(sec *security.Security) error {
	url := "https://finnhub.io/api/v1/quote?symbol=" + sec.Ticker

	response, err := fetch(url)
	if err != nil {
		return fmt.Errorf("error fetching quote %s %s", sec.Ticker, err)
	}

	err = parseQuote(response, sec)
	if err != nil {
		return fmt.Errorf("error parsing quote %s", err)
	}

	return nil
}

func getMetric(sec *security.Security) error {
	url := "https://finnhub.io/api/v1/stock/metric?metric=all&symbol=" + sec.Ticker

	response, err := fetch(url)
	if err != nil {
		return fmt.Errorf("error fetching metric %s %s", sec.Ticker, err)
	}

	err = parseMetric(response, sec)
	if err != nil {
		return fmt.Errorf("error parsing metric %s", err)
	}

	return nil
}

// GetStock looks up a single ticker symbol returns the sec
func GetStock(sec *security.Security) error {
	err := getQuote(sec)
	if err != nil {
		return err
	}

	return getMetric(sec)
}
