package urlscan

import (
	"bytes"
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	Name = "urlscan"
)

var _ providers.Provider = (*Client)(nil)

type Client struct {
	config *providers.Config
}

func New(c *providers.Config) *Client {
	if c.URLScan.Host != "" {
		setBaseURL(c.URLScan.Host)
	}

	return &Client{config: c}
}

func (c *Client) Name() string {
	return Name
}
func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
	var searchAfter string
	var header httpclient.Header

	if c.config.URLScan.APIKey != "" {
		header.Key = "API-Key"
		header.Value = c.config.URLScan.APIKey
	}

	page := 0
paginate:
	for {
		select {
		case <-ctx.Done():
			break paginate
		default:
			if c.config.Verbose {
				logrus.WithFields(logrus.Fields{"provider": Name, "page": page}).Infof("fetching %s", domain)
			}
			apiURL := c.formatURL(domain, searchAfter)
			resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout, header)
			if err != nil {
				return fmt.Errorf("failed to fetch urlscan: %s", err)
			}
			var result apiResponse
			decoder := jsoniter.NewDecoder(bytes.NewReader(resp))
			decoder.UseNumber()
			if err = decoder.Decode(&result); err != nil {
				return fmt.Errorf("failed to decode urlscan result:  %s", err)
			}
			// rate limited
			if result.Status == 429 {
				if c.config.Verbose {
					logrus.WithField("provider", "urlscan").Warnf("urlscan responded with 429")
				}
				break paginate
			}

			total := len(result.Results)
			for i, res := range result.Results {
				if res.Page.Domain == domain || (c.config.IncludeSubdomains && strings.HasSuffix(res.Page.Domain, domain)) {
					results <- res.Page.URL
				}

				if i == total-1 {
					sortParam := parseSort(res.Sort)
					if sortParam != "" {
						searchAfter = sortParam
					} else {
						break paginate
					}
				}
			}

			if !result.HasMore {
				break paginate
			}
			page++
		}
	}
	return nil
}

func (c *Client) formatURL(domain string, after string) string {
	if after != "" {
		return fmt.Sprintf(_BaseURL+"api/v1/search/?q=domain:%s&size=100", domain) + "&search_after=" + after
	}

	return fmt.Sprintf(_BaseURL+"api/v1/search/?q=domain:%s&size=100", domain)
}

func setBaseURL(baseURL string) {
	_BaseURL = baseURL
}
