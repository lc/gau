package otx

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bobesa/go-domain-util/domainutil"
	jsoniter "github.com/json-iterator/go"
	"github.com/lc/gau/v2/pkg/httpclient"
	"github.com/lc/gau/v2/pkg/providers"
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
	numThreads := c.config.ProviderThreads
	if numThreads == 0 {
		numThreads = 3
	}

	pageChan := make(chan uint, numThreads)
	var wg sync.WaitGroup
	var fetchErr error
	var errMu sync.Mutex
	var stopOnce sync.Once
	stopCh := make(chan struct{})

	// Page dispatcher: sequentially increments pages, stops when receiving stop signal
	go func() {
		defer close(pageChan)
		for page := uint(1); ; page++ {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case pageChan <- page:
			}
		}
	}()

	// Workers: fetch pages from pageChan
	for i := uint(0); i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range pageChan {
				select {
				case <-ctx.Done():
					return
				default:
				}
				logrus.WithFields(logrus.Fields{"provider": Name, "page": p - 1}).Infof("fetching %s", domain)
				apiURL := c.formatURL(domain, p)
				resp, err := httpclient.MakeRequest(c.config.Client, apiURL, c.config.MaxRetries, c.config.Timeout)
				if err != nil {
					var statusErr *httpclient.StatusCodeError
					if errors.As(err, &statusErr) {
						logrus.WithFields(logrus.Fields{
							"provider": Name,
							"domain":   domain,
							"page":     p - 1,
							"status":   statusErr.Code,
							"error":    statusErr.Error(),
						}).Warn("OTX HTTP error")
						if statusErr.Code == 429 {
							errMu.Lock()
							if fetchErr == nil {
								fetchErr = fmt.Errorf("OTX rate limited at page %d", p)
							}
							errMu.Unlock()
							stopOnce.Do(func() { close(stopCh) })
							return
						}
					} else {
						logrus.WithFields(logrus.Fields{
							"provider": Name,
							"domain":   domain,
							"page":     p - 1,
							"error":    err.Error(),
						}).Warn("failed to fetch OTX")
						errMu.Lock()
						if fetchErr == nil {
							fetchErr = fmt.Errorf("failed to fetch alienvault(%d): %s", p, err)
						}
						errMu.Unlock()
					}
					continue
				}
				var result otxResult
				if err := jsoniter.Unmarshal(resp, &result); err != nil {
					errMu.Lock()
					if fetchErr == nil {
						fetchErr = fmt.Errorf("failed to decode otx results for page %d: %s", p, err)
					}
					errMu.Unlock()
					continue
				}

				for _, entry := range result.URLList {
					select {
					case <-ctx.Done():
						return
					case results <- entry.URL:
					}
				}

				if !result.HasNext {
					stopOnce.Do(func() { close(stopCh) })
					return
				}
			}
		}()
	}

	wg.Wait()
	return fetchErr
}

func (c *Client) formatURL(domain string, page uint) string {
	category := "hostname"
	if !domainutil.HasSubdomain(domain) {
		category = "domain"
	}
	if domainutil.HasSubdomain(domain) && c.config.IncludeSubdomains {
		domain = domainutil.Domain(domain)
		category = "domain"
	}

	return fmt.Sprintf("%sapi/v1/indicators/%s/%s/url_list?limit=100&page=%d", _BaseURL, category, domain, page)
}

var _BaseURL = "https://otx.alienvault.com/"

func setBaseURL(baseURL string) {
	_BaseURL = baseURL
}
