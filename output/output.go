package output

import (
	"bufio"
	"io"
	"net/url"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type JSONResult struct {
	Url string `json:"url"`
}
func WriteURLs(results <-chan string, writer io.Writer, blacklistMap map[string]struct{}) error {
	wr := bufio.NewWriter(writer)
	str := &strings.Builder{}
	for result := range results {
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result)
			if err != nil {
				continue
			}
			base := strings.Split(path.Base(u.Path),".")
			ext := base[len(base)-1]
			if ext != "" {
				_, ok := blacklistMap[strings.ToLower(ext)]
				if ok {
					continue
				}
			}
		}
		str.WriteString(result)
		str.WriteRune('\n')
		_, err := wr.WriteString(str.String())
		if err != nil {
			wr.Flush()
			return err
		}
		str.Reset()
	}
	return wr.Flush()
}
func WriteURLsJSON(results <-chan string, writer io.Writer, blacklistMap map[string]struct{}) {
	var jr JSONResult
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result)
			if err != nil {
				continue
			}
			base := strings.Split(path.Base(u.Path),".")
			ext := base[len(base)-1]
			if ext != "" {
				_, ok := blacklistMap[strings.ToLower(ext)]
				if ok {
					continue
				}
			}
		}
		jr.Url = result
		enc.Encode(jr)
	}
}
