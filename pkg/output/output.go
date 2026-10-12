package output

import (
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/bytebufferpool"
)

type JSONResult struct {
	Url string `json:"url"`
}

func WriteURLs(writer io.Writer, results <-chan string, blacklistMap mapset.Set[string], RemoveParameters bool, urlCount *int64) error {
	lastURL := mapset.NewThreadUnsafeSet[string]()
	for result := range results {
		buf := bytebufferpool.Get()
		u, err := url.Parse(result)
		if err != nil {
			continue
		}
		ext := strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), ".")
		if ext != "" && blacklistMap.Contains(ext) {
			continue
		}

		if RemoveParameters {
			if lastURL.Contains(u.Host + u.Path) {
				continue // already seen this endpoint, skip duplicate params
			}
			lastURL.Add(u.Host + u.Path)
		}

		buf.B = append(buf.B, []byte(result)...)
		buf.B = append(buf.B, "\n"...)
		_, err = writer.Write(buf.B)
		if err != nil {
			return err
		}
		atomic.AddInt64(urlCount, 1)
		// Real-time flush: sync stdout after each write to prevent data loss
		if writer == os.Stdout {
			os.Stdout.Sync()
		}
		bytebufferpool.Put(buf)
	}
	return nil
}

func WriteURLsJSON(writer io.Writer, results <-chan string, blacklistMap mapset.Set[string], RemoveParameters bool, urlCount *int64) {
	var jr JSONResult
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		u, err := url.Parse(result)
		if err != nil {
			continue
		}
		ext := strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), ".")
		if ext != "" && blacklistMap.Contains(ext) {
			continue
		}
		jr.Url = result
		if err := enc.Encode(jr); err != nil {
			// todo: handle this error
			continue
		}
		atomic.AddInt64(urlCount, 1)
		// Real-time flush: sync stdout after each write to prevent data loss
		if writer == os.Stdout {
			os.Stdout.Sync()
		}
	}
}
