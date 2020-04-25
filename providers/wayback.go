package providers

import (
	"encoding/json"
	"fmt"
)

type WaybackProvider struct {
	*Config
}

type WaybackResult [][]string

const waybackResultsLimit = 200

func (w *WaybackProvider) formatURL(domain string, page int) string {
	if w.IncludeSubdomains {
		domain = "*." + domain
	}

	return fmt.Sprintf(
		"http://web.archive.org/cdx/search/cdx?url=%s/*&output=json&collapse=urlkey&fl=original&limit=%d&offset=%d",
		domain, waybackResultsLimit, page*waybackResultsLimit,
	)
}

func (w *WaybackProvider) Fetch(domain string, results chan<- string) error {
	for page := 0; ; page++ {
		resp, err := w.MakeRequest(w.formatURL(domain, page))
		if err != nil {
			return err
		}

		var result WaybackResult
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return err
		}

		_ = resp.Body.Close()

		for i, entry := range result {
			// Skip first result by default
			if i != 0 {
				results <- entry[0]
			}
		}

		if len(result) < waybackResultsLimit {
			break
		}
	}

	return nil
}
