package otx

import (
	"context"
	"fmt"
	"github.com/bobesa/go-domain-util/domainutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/pkg/httpclient"
	"github.com/lc/gau/pkg/providers"
	"github.com/sirupsen/logrus"
)

const (
	Name = "otx"
)

type Client struct {
	config *providers.Config
}

var _ providers.Provider = (*Client)(nil)

func New(c *providers.Config) *Client {
	if c.OTX != "" {
		setBaseURL(c.OTX)
	}
	return &Client{config: c}
}

type otxResult struct {
	HasNext    bool `json:"has_next"`
	ActualSize int  `json:"actual_size"`
	URLList    []struct {
		Domain   string `json:"domain"`
		URL      string `json:"url"`
		Hostname string `json:"hostname"`
		HTTPCode int    `json:"httpcode"`
		PageNum  int    `json:"page_num"`
		FullSize int    `json:"full_size"`
		Paged    bool   `json:"paged"`
	} `json:"url_list"`
}

func (c *Client) Name() string {
	return Name
}

func (c *Client) Fetch(ctx context.Context, domain string, results chan string) error {
paginate:
	for page := 1; ; page++ {
		select {
		case <-ctx.Done():
			break paginate
		default:
			if c.config.Verbose {
				logrus.WithFields(logrus.Fields{"provider": Name, "page": page - 1}).Infof("fetching %s", domain)
			}
			apiURL := c.formatURL(domain, page)
			resp, err := httpclient.MakeRequest(c.config.Client, apiURL, int(c.config.MaxRetries))
			if err != nil {
				return fmt.Errorf("failed to fetch alienvault(%d): %s", page, err)
			}
			var result otxResult
			if err := jsoniter.Unmarshal(resp, &result); err != nil {
				return fmt.Errorf("failed to decode otx results for page %d: %s", page, err)
			}

			for _, entry := range result.URLList {
				results <- entry.URL
			}

			if !result.HasNext {
				break paginate
			}
		}
	}
	return nil
}

func (c *Client) formatURL(domain string, page int) string {
	if !domainutil.HasSubdomain(domain) {
		return fmt.Sprintf(_BaseURL+"api/v1/indicators/domain/%s/url_list?limit=100&page=%d",
			domain, page,
		)
	} else if domainutil.HasSubdomain(domain) && c.config.IncludeSubdomains {
		return fmt.Sprintf(_BaseURL+"api/v1/indicators/domain/%s/url_list?limit=100&page=%d",
			domainutil.Domain(domain), page,
		)
	} else {
		return fmt.Sprintf(_BaseURL+"api/v1/indicators/hostname/%s/url_list?limit=100&page=%d",
			domain, page,
		)
	}
}

var _BaseURL = "https://otx.alienvault.com/"

func setBaseURL(baseURL string) {
	_BaseURL = baseURL
}
