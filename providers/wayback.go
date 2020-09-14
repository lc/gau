package providers

import (
	"encoding/json"
	"fmt"
	"time"
)

type WaybackProvider struct {
	*Config
}

type WaybackPaginationResult uint

type WaybackResult [][]string

func NewWaybackProvider(config *Config) Provider {
	return &WaybackProvider{Config: config}
}

func (w *WaybackProvider) formatURL(domain string, page uint) string {
	if w.IncludeSubdomains {
		domain = "*." + domain
	}

	return fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=%s/*&output=json&collapse=urlkey&fl=original&page=%d",
		domain, page,
	)
}

// Fetch the number of pages.
func (w *WaybackProvider) getPagination(domain string) (WaybackPaginationResult, error) {
	url := fmt.Sprintf("%s&showNumPages=true", w.formatURL(domain, 0))

	resp, err := w.MakeRequest(url)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var paginationResult WaybackPaginationResult
	if err = json.NewDecoder(resp.Body).Decode(&paginationResult); err != nil {
		return 0, err
	}

	time.Sleep(time.Millisecond * 100)
	return paginationResult, nil
}

func (w *WaybackProvider) Fetch(domain string, results chan<- string) error {
	pages, err := w.getPagination(domain)
	if err != nil {
		return fmt.Errorf("failed to fetch wayback pagination: %s", err)
	}

	for page := uint(0); page < uint(pages); page++ {
		resp, err := w.MakeRequest(w.formatURL(domain, page))
		if err != nil {
			return fmt.Errorf("failed to fetch wayback results page %d: %s", page, err)
		}

		var result WaybackResult
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to decode wayback results for page %d: %s", page, err)
		}

		_ = resp.Body.Close()
		for i, entry := range result {
			// Skip first result by default
			if i != 0 {
				results <- entry[0]
			}
		}
	}

	return nil
}
