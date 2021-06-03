package providers

import (
	"encoding/json"
	"fmt"
	"github.com/bobesa/go-domain-util/domainutil"
	"strings"
)

type OTXProvider struct {
	*Config
}

type OTXResult struct {
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

const otxResultsLimit = 200

func NewOTXProvider(config *Config) Provider {
	return &OTXProvider{Config: config}
}

func (o *OTXProvider) formatURL(domain string, page int) string {
	if !domainutil.HasSubdomain(domain) {
		return fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?limit=%d&page=%d",
			domain, otxResultsLimit, page,
		)
	} else if domainutil.HasSubdomain(domain) && o.IncludeSubdomains {
		return fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/url_list?limit=%d&page=%d",
			domainutil.Domain(domain), otxResultsLimit, page,
		)
	} else {
		return fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/hostname/%s/url_list?limit=%d&page=%d",
			domain, otxResultsLimit, page,
		)
	}
}

func (o *OTXProvider) Fetch(domain string, results chan<- string) error {
	for page := 0; ; page++ {
		resp, err := o.MakeRequest(o.formatURL(domain, page))
		if err != nil {
			return fmt.Errorf("failed to fetch otx results page %d: %s", page, err)
		}

		var result OTXResult
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to decode otx results for page %d: %s", page, err)
		}

		_ = resp.Body.Close()

		for _, entry := range result.URLList {
			if o.IncludeSubdomains {
				if !domainutil.HasSubdomain(domain) {
					results <- entry.URL
				} else {
					if strings.Contains(strings.ToLower(entry.Hostname), strings.ToLower(domain)) {
						results <- entry.URL
					}
				}
			} else {
				if strings.EqualFold(domain, entry.Hostname) {
					results <- entry.URL
				}
			}
		}

		if !result.HasNext {
			break
		}
	}

	return nil
}
