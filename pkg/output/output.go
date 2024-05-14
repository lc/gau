package output

import (
	"io"
	"net/url"
	"path"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/bytebufferpool"
)

type JSONResult struct {
	Url string `json:"url"`
}

func Blacklisted(u *url.URL, blacklistMap mapset.Set[string], blacklistpathsMap mapset.Set[string]) bool {
	if path.Ext(u.Path) != "" {
		if blacklistMap.Contains(strings.ToLower(path.Ext(u.Path))) || blacklistMap.Contains(strings.ToLower(path.Ext(u.RawQuery))) {
			return true
		}
		for path := range blacklistpathsMap.Iter() {
			if strings.Contains(u.Path, path) {
				return true
			}
		}
	}
	return false
}

func WriteURLs(writer io.Writer, results <-chan string, blacklistMap mapset.Set[string], blacklistpathsMap mapset.Set[string], RemoveParameters bool) error {
	lastURL := mapset.NewThreadUnsafeSet[string]()
	for result := range results {
		buf := bytebufferpool.Get()
		u, err := url.Parse(result)
		if err != nil {
			continue
		}

		if Blacklisted(u, blacklistMap, blacklistpathsMap) {
			continue
		}

		if RemoveParameters && !lastURL.Contains(u.Host+u.Path) {
			continue
		}
		lastURL.Add(u.Host + u.Path)

		buf.B = append(buf.B, []byte(result)...)
		buf.B = append(buf.B, "\n"...)
		_, err = writer.Write(buf.B)
		if err != nil {
			return err
		}
		bytebufferpool.Put(buf)
	}
	return nil
}

func WriteURLsJSON(writer io.Writer, results <-chan string, blacklistMap mapset.Set[string], blacklistpathsMap mapset.Set[string], RemoveParameters bool) {
	var jr JSONResult
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		u, err := url.Parse(result)
		if err != nil {
			continue
		}
		if Blacklisted(u, blacklistMap, blacklistpathsMap) {
			continue
		}
		jr.Url = result
		if err := enc.Encode(jr); err != nil {
			// todo: handle this error
			continue
		}
	}
}
