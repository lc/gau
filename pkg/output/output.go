package output

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/bytebufferpool"
	"io"
	"net/url"
	"path"
	"strings"
)

type JSONResult struct {
	Url string `json:"url"`
}

func WriteURLs(writer io.Writer, results <-chan string, blacklistMap map[string]struct{}) error {
	for result := range results {
		buf := bytebufferpool.Get()
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result)
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
		buf.B = append(buf.B, []byte(result)...)
		buf.B = append(buf.B, "\n"...)
		_, err := writer.Write(buf.B)
		if err != nil {
			return err
		}
		bytebufferpool.Put(buf)
	}
	return nil
}

func WriteURLsJSON(writer io.Writer, results <-chan string, blacklistMap map[string]struct{}) {
	var jr JSONResult
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		if len(blacklistMap) != 0 {
			u, err := url.Parse(result)
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
		jr.Url = result
		if err := enc.Encode(jr); err != nil {
			// todo: handle this error
			continue
		}
	}
}
