package output

import (
	"io"
	"net/url"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/bytebufferpool"
)

// Result of lookup from providers.
type Result struct {
	URL      string `json:"url,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func WriteURLs(writer io.Writer, results <-chan Result, blacklistMap map[string]struct{}, RemoveParameters bool) error {
	lastURL := make(map[string]struct{})
	for result := range results {
		buf := bytebufferpool.Get()
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result.URL)
			if err != nil {
				continue
			}
			base := strings.Split(path.Base(u.Path), ".")
			ext := base[len(base)-1]
			if ext != "" {
				_, ok := blacklistMap[strings.ToLower(ext)]
				if ok {
					continue
				}
			}
		}
		if RemoveParameters {
			u, err := url.Parse(result.URL)
			if err != nil {
				continue
			}
			if _, ok := lastURL[u.Host+u.Path]; ok {
				continue
			} else {
				lastURL[u.Host+u.Path] = struct{}{}
			}

		}

		buf.B = append(buf.B, []byte(result.URL)...)
		buf.B = append(buf.B, "\n"...)
		_, err := writer.Write(buf.B)
		if err != nil {
			return err
		}
		bytebufferpool.Put(buf)
	}
	return nil
}

func WriteURLsJSON(writer io.Writer, results <-chan Result, blacklistMap map[string]struct{}, RemoveParameters bool) {
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result.URL)
			if err != nil {
				continue
			}
			base := strings.Split(path.Base(u.Path), ".")
			ext := base[len(base)-1]
			if ext != "" {
				_, ok := blacklistMap[strings.ToLower(ext)]
				if ok {
					continue
				}
			}
		}
		if err := enc.Encode(result); err != nil {
			// todo: handle this error
			continue
		}
	}
}
