package urlscan

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "urlscan"
)

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

	for page := uint(0); ; page++ {
		select {
		case <-ctx.Done():
			return nil
		default:
			logrus.WithFields(logrus.Fields{"provider": Name, "page": page}).Infof("fetching %s", domain)
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
				logrus.WithField("provider", "urlscan").Warnf("urlscan responded with 429, probably being rate limited")
				return nil
			}

			total := len(result.Results)
			for i, res := range result.Results {
				if res.Page.Domain == domain || (c.config.IncludeSubdomains && strings.HasSuffix(res.Page.Domain, domain)) {
					results <- res.Page.URL
				}

				if i == total-1 {
					sortParam := parseSort(res.Sort)
					if sortParam == "" {
						return nil
					}
					searchAfter = sortParam
				}
			}

			if !result.HasMore {
				return nil
			}
		}
	}
}

func (c *Client) formatURL(domain string, after string) string {
	if after != "" {
		after = "&search_after=" + after
	}

	return fmt.Sprintf(_BaseURL+"api/v1/search/?q=domain:%s&size=100", domain) + after
}

func setBaseURL(baseURL string) {
	_BaseURL = baseURL
}
