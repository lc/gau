package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"
	"math/rand"
)

type CommonProvider struct {
	*Config
	apiURL string
}

type CommonResult struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type CommonPaginationResult struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type CommonAPIResult []struct {
	API string `json:"cdx-api"`
}

func NewCommonProvider(config *Config) (Provider, error) {
	c := CommonProvider{Config: config}

	// Fetch the list of available CommonCrawl Api URLs.
	resp, err := c.MakeRequest("http://index.commoncrawl.org/collinfo.json")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var apiResult CommonAPIResult
	if err = json.NewDecoder(resp.Body).Decode(&apiResult); err != nil {
		return nil, err
	}

	c.apiURL = apiResult[0].API
	return &c, nil
}

func (c *CommonProvider) formatURL(domain string, page uint) string {
	if c.IncludeSubdomains {
		domain = "*." + domain
	}

	return fmt.Sprintf("%s?url=%s/*&output=json&fl=url&page=%d", c.apiURL, domain, page)
}

// Fetch the number of pages.
func (c *CommonProvider) getPagination(domain string) (*CommonPaginationResult, error) {
	url := fmt.Sprintf("%s&showNumPages=true", c.formatURL(domain, 0))

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
	pagination, err := c.getPagination(domain)
	if err != nil {
		return fmt.Errorf("failed to fetch common pagination: %s", err)
	}

	for page := uint(0); page < pagination.Pages; page++ {
		resp, err := c.MakeRequest(c.formatURL(domain, page))
		if err != nil {
			return fmt.Errorf("failed to fetch common results page %d: %s", page, err)
		}

		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			var result CommonResult
			if err := json.Unmarshal(sc.Bytes(), &result); err != nil {
				_ = resp.Body.Close()
				return fmt.Errorf("failed to decode common results for page %d: %s", page, err)
			}

			if result.Error != "" {
				return fmt.Errorf("received an error from common api: %s", result.Error)
			}

			results <- result.URL
		}

		_ = resp.Body.Close()
	}

	return nil
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}