package wayback

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/output"
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

func New(c *providers.Config, filters providers.Filters) *Client {
	return &Client{
		filters: filters,
		config:  c,
	}
}

func (c *Client) Name() string {
	return Name
}

// waybackResult holds the response from the wayback API
type waybackResult [][]string

// Fetch fetches all urls for a given domain and sends them to a channel.
// It returns an error should one occur.
func (c *Client) Fetch(ctx context.Context, domain string, results chan output.Result) error {
	pages, err := c.getPagination(domain)
	if err != nil {
		return fmt.Errorf("failed to fetch wayback pagination: %s", err)
	}
	for page := uint(0); page < pages; page++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			if c.config.Verbose {
				logrus.WithFields(logrus.Fields{"provider": Name, "page": page}).Infof("fetching %s", domain)
			}
			apiURL := c.formatURL(domain, page)
			// make HTTP request
			resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
			if err != nil {
				return fmt.Errorf("failed to fetch wayback results page %d: %s", page, err)
			}

			var result waybackResult
			if err = jsoniter.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to decode wayback results for page %d: %s", page, err)
			}

			// check if there's results, wayback's pagination response
			// is not always correct when using a filter
			if len(result) == 0 {
				break
			}

			// output results
			for i, entry := range result {
				// Skip first result by default
				if i != 0 {
					results <- output.Result{
						URL:      entry[0],
						Provider: Name,
					}
				}
			}
		}
	}
	return nil
}

// formatUrl returns a formatted URL for the Wayback API
func (c *Client) formatURL(domain string, page uint) string {
	if c.config.IncludeSubdomains {
		domain = "*." + domain
	}
	filterParams := c.filters.GetParameters(true)
	return fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=%s/*&output=json&collapse=urlkey&fl=original&page=%d",
		domain, page,
	) + filterParams
}

// getPagination returns the number of pages for Wayback
func (c *Client) getPagination(domain string) (uint, error) {
	url := fmt.Sprintf("%s&showNumPages=true", c.formatURL(domain, 0))
	resp, err := httpclient.MakeRequest(c.config.Client, url, c.config.MaxRetries, c.config.Timeout)

	if err != nil {
		return 0, err
	}

	var paginationResult uint

	if err = jsoniter.Unmarshal(resp, &paginationResult); err != nil {
		return 0, err
	}

	return paginationResult, nil
}
