package providers

import (
	"math/rand"
	"net/http"
	"time"
)

const (
	// Version of gau
	Version = `1.1.0`
)

// A generic interface for providers
type Provider interface {
	Fetch(string, chan<- string) error
}
type Config struct {
	Verbose           bool
	MaxRetries        uint
	IncludeSubdomains bool
	Client            *http.Client
	Providers         []string
	Blacklist         map[string]struct{}
	Output            string
	JSON              bool
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}

// MakeRequest tries to make a GET request for the given URL and retries on failure.
func (c *Config) MakeRequest(url string) (resp *http.Response, err error) {
	for retries := int(c.MaxRetries); ; retries-- {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", getUserAgent())

		resp, err = c.Client.Do(req)
		if err != nil {
			if retries == 0 {
				return nil, err
			}

			continue
		}

		break
	}

	return
}
