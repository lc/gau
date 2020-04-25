package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
)

type CommonProvider struct {
	*Config
}

type CommonResult struct {
	URL string `json:"url"`
}

type CommonPaginationResult struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type CommonAPIResult []struct {
	API string `json:"cdx-api"`
}

func (c *CommonProvider) formatURL(apiURL string, domain string, page uint) string {
	if c.IncludeSubdomains {
		domain = "*." + domain
	}

	return fmt.Sprintf("%s?url=%s/*&output=json&fl=url&page=%d", apiURL, domain, page)
}

// Fetch the list of available CommonCrawl Api URLs.
func (c *CommonProvider) getAPIs() (CommonAPIResult, error) {
	resp, err := c.MakeRequest("http://index.commoncrawl.org/collinfo.json")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var apiResult CommonAPIResult
	if err = json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		return nil, err
	}

	return apiResult, nil
}

// Fetch the number of pages.
func (c *CommonProvider) getPagination(apiURL string, domain string) (*CommonPaginationResult, error) {
	url := fmt.Sprintf("%s&showNumPages=true", c.formatURL(apiURL, domain, 0))

	resp, err := c.MakeRequest(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var paginationResult CommonPaginationResult
	if err = json.NewDecoder(resp.Body).Decode(&paginationResult); err != nil {
		return nil, err
	}

	return &paginationResult, nil
}

func (c *CommonProvider) Fetch(domain string, results chan<- string) error {
	api, err := c.getAPIs()
	if err != nil {
		return err
	}

	pagination, err := c.getPagination(api[0].API, domain)
	if err != nil {
		return err
	}

	for page := uint(0); page < pagination.Pages; page++ {
		resp, err := c.MakeRequest(c.formatURL(api[0].API, domain, page))
		if err != nil {
			return err
		}

		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			var result CommonResult
			if err := json.Unmarshal(sc.Bytes(), &result); err != nil {
				_ = resp.Body.Close()
				return err
			}
			results <- result.URL
		}

		_ = resp.Body.Close()
	}

	return nil
}
