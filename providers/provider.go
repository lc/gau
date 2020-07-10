package providers

import (
	"net/http"
)

const (
	// Version of gau
	Version = `1.0.3`
	// UserAgent for the HTTP Client
	userAgent = "Mozilla/5.0 (compatible; gau/" + Version + "; https://github.com/lc/gau)"
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
	Output            string
	JSON              bool
}

// MakeRequest tries to make a GET request for the given URL and retries on failure.
func (c *Config) MakeRequest(url string) (resp *http.Response, err error) {
	for retries := int(c.MaxRetries); ; retries-- {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("User-Agent", userAgent)
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
