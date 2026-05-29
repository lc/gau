package commoncrawl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "commoncrawl"
)

// verify interface compliance
var _ providers.Provider = (*Client)(nil)

// Client is the structure that holds the Filters and the Client's configuration
type Client struct {
	filters providers.Filters
	config  *providers.Config

	apiURL string
}

func New(c *providers.Config, filters providers.Filters) (*Client, error) {
	// Fetch the list of available CommonCrawl Api URLs.
	resp, err := httpclient.MakeRequest(c.Client, "http://index.commoncrawl.org/collinfo.json", c.MaxRetries, c.Timeout)
	if err != nil {
		return nil, err
	}

	var r apiResult
	if err = jsoniter.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	if len(r) == 0 {
		return nil, errors.New("failed to grab latest commoncrawl index")
	}

	return &Client{config: c, filters: filters, apiURL: r[0].API}, nil
}

func (c *Client) Name() string {
	return Name
}

// Fetch fetches all urls for a given domain and sends them to a channel.
// It returns an error should one occur.
func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	p, err := c.getPagination(domain)
	if err != nil {
		return err
	}
	// 0 pages means no results
	if p.Pages == 0 {
		logrus.WithFields(logrus.Fields{"provider": Name}).Infof("no results for %s", domain)
		return nil
	}

	for page := uint(0); page < p.Pages; page++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			logrus.WithFields(logrus.Fields{"provider": Name, "page": page}).Infof("fetching %s", domain)
			apiURL := c.formatURL(domain, page)
			resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
			if err != nil {
				return fmt.Errorf("failed to fetch commoncrawl(%d): %s", page, err)
			}

			sc := bufio.NewScanner(bytes.NewReader(resp))
			for sc.Scan() {
				var res apiResponse
				if err := jsoniter.Unmarshal(sc.Bytes(), &res); err != nil {
					return fmt.Errorf("failed to decode commoncrawl result:  %s", err)
				}
				if res.Error != "" {
					return fmt.Errorf("received an error from commoncrawl: %s", res.Error)
				}

				results <- res.URL
			}
		}
	}
	return nil
}

func (c *Client) formatURL(domain string, page uint) string {
	if c.config.IncludeSubdomains {
		domain = "*." + domain
	}

	filterParams := c.filters.GetParameters(false)

	return fmt.Sprintf("%s?url=%s/*&output=json&fl=url&page=%d", c.apiURL, domain, page) + filterParams
}

// Fetch the number of pages.
func (c *Client) getPagination(domain string) (r paginationResult, err error) {
	url := fmt.Sprintf("%s&showNumPages=true", c.formatURL(domain, 0))
	var resp []byte

	resp, err = httpclient.MakeRequest(c.config.Client, url, c.config.MaxRetries, c.config.Timeout)
	if err != nil {
		return
	}

	err = jsoniter.Unmarshal(resp, &r)
	return
}
