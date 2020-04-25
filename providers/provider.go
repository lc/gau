package providers

import (
	"net/http"
)

// A generic interface for providers
type Provider interface {
	Fetch(string, chan<- string) error
}

type Config struct {
	MaxRetries        uint
	IncludeSubdomains bool
	Client            *http.Client
}

// MakeRequest tries to make a GET request for the given URL and retries on failure.
func (c *Config) MakeRequest(url string) (resp *http.Response, err error) {
	for retries := int(c.MaxRetries); ; retries-- {
		resp, err = c.Client.Get(url)
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
