package wayback

import (
	"context"
	"errors"
	"fmt"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "wayback"
)

// verify interface compliance
var _ providers.Provider = (*Client)(nil)

// Client is the structure that holds the WaybackFilters and the Client's configuration
type Client struct {
	filters providers.Filters
	config  *providers.Config
}

func New(config *providers.Config, filters providers.Filters) *Client {
	return &Client{filters, config}
}

func (c *Client) Name() string {
	return Name
}

// waybackResult holds the response from the wayback API
type waybackResult [][]string

// Fetch fetches all urls for a given domain and sends them to a channel.
// It returns an error should one occur.
func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	// Use provider threads for concurrent pagination, default to 3 if not set
	numThreads := c.config.ProviderThreads
	if numThreads == 0 {
		numThreads = 3
	}

	// channel to collect errors from goroutines
	var fetchErr error
	var errMu sync.Mutex
	var stopOnce sync.Once
	stopCh := make(chan struct{})

	// pageChan is a buffered channel for page dispatching
	pageChan := make(chan uint, numThreads)

	// Page dispatcher: sequentially increments pages, stops when receiving stop signal
	go func() {
		defer close(pageChan)
		for page := uint(0); ; page++ {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case pageChan <- page:
			}
		}
	}()

	// Workers: fetch pages from pageChan, notify dispatcher to stop on empty results
	var wg sync.WaitGroup
	for i := uint(0); i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range pageChan {
				select {
				case <-ctx.Done():
					return
				default:
				}
				logrus.WithFields(logrus.Fields{"provider": Name, "page": page}).Infof("fetching %s", domain)
				apiURL := c.formatURL(domain, page)
				// make HTTP request
				resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
				if err != nil {
					var statusErr *httpclient.StatusCodeError
					if errors.As(err, &statusErr) && statusErr.Code == 400 {
						logrus.WithFields(logrus.Fields{
							"provider": Name,
							"domain":   domain,
							"page":     page,
							"status":   400,
						}).Info("Wayback: no more pages")
						stopOnce.Do(func() { close(stopCh) })
						return
					}
					logrus.WithFields(logrus.Fields{
						"provider": Name,
						"domain":   domain,
						"page":     page,
						"error":    err.Error(),
					}).Warn("failed to fetch wayback")
					errMu.Lock()
					if fetchErr == nil {
						fetchErr = fmt.Errorf("failed to fetch wayback results page %d: %s", page, err)
					}
					errMu.Unlock()
					continue
				}
				var result waybackResult
				if err = jsoniter.Unmarshal(resp, &result); err != nil {
					errMu.Lock()
					if fetchErr == nil {
						fetchErr = fmt.Errorf("failed to decode wayback results for page %d: %s", page, err)
					}
					errMu.Unlock()
					continue
				}

				// check if there's results, wayback's pagination response
				// is not always correct when using a filter
				if len(result) == 0 {
					// Notify dispatcher to stop when empty result is encountered
					stopOnce.Do(func() { close(stopCh) })
					return
				}

				// output results
				// Slicing as [1:] to skip first result by default
				for _, entry := range result[1:] {
					select {
					case <-ctx.Done():
						return
					case results <- entry[0]:
					}
				}
			}
		}()
	}

	wg.Wait()

	return fetchErr
}

// formatUrl returns a formatted URL for the Wayback API
func (c *Client) formatURL(domain string, page uint) string {
	if c.config.IncludeSubdomains {
		domain = "*." + domain
	}
	filterParams := c.filters.GetParameters(true)
	return fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=%s/*&output=json&collapse=urlkey&fl=original&pageSize=100&page=%d",
		domain, page,
	) + filterParams
}
