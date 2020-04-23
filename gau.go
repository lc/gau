package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type OTXResult struct {
	HasNext    bool `json:"has_next"`
	ActualSize int  `json:"actual_size"`
	URLList    []struct {
		Domain   string `json:"domain"`
		URL      string `json:"url"`
		Hostname string `json:"hostname"`
		Httpcode int    `json:"httpcode"`
		PageNum  int    `json:"page_num"`
		FullSize int    `json:"full_size"`
		Paged    bool   `json:"paged"`
	} `json:"url_list"`
}
type CommonCrawlInfo []struct {
	CdxAPI string `json:"cdx-api"`
}

var (
	IncludeSubs bool
	MaxRetries  int
	client      = &http.Client{
		Timeout: time.Second * 15,
	}
	CCApi string
)

func main() {
	var domains []string
	flag.BoolVar(&IncludeSubs, "subs", false, "include subdomains of target domain")
	flag.IntVar(&MaxRetries, "retries", 5, "amount of retries for http client")
	flag.Parse()
	if flag.NArg() > 0 {
		domains = []string{flag.Arg(0)}
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			domains = append(domains, s.Text())
		}
	}
	var err error
	CCApi, err = getCurrentCC()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	for _, domain := range domains {
		Run(domain)
	}
}

type fetch func(string) ([]string, error)

func Run(domain string) {
	fetchers := []fetch{getWaybackUrls, getCommonCrawlURLs, getOtxUrls}
	for _, fn := range fetchers {
		found, err := fn(domain)
		if err != nil {
			if strings.Contains(err.Error(), "cc api") {
				continue
			}
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}
		for _, f := range found {
			fmt.Println(f)
		}
	}
}
func getOtxUrls(hostname string) ([]string, error) {
	var urls []string
	page := 0
	retries := MaxRetries
	for {
		var o = &OTXResult{}
		for retries > 0 {
			r, err := client.Get(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/hostname/%s/url_list?limit=50&page=%d", hostname, page))
			if err != nil {
				retries -= 1
				if retries == 0 {
					return nil, errors.New(fmt.Sprintf("http request to OTX failed: %s", err.Error()))
				}

			}
			err = json.NewDecoder(r.Body).Decode(o)
			if err != nil {
				retries -= 1
				if retries == 0 {
					return nil, errors.New(fmt.Sprintf("error in parsing JSON from alienvault: %s", err.Error()))
				}
				continue
			}
			r.Body.Close()
			break

		}
		for _, url := range o.URLList {
			urls = append(urls, url.URL)
		}
		if !o.HasNext {
			break
		}
		page++
	}
	return urls, nil
}
func getWaybackUrls(hostname string) ([]string, error) {
	wildcard := "*."
	var waybackresp [][]string
	if !IncludeSubs {
		wildcard = ""
	}
	retries := MaxRetries
	var found []string
	for retries > 0 {
		tg := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey&fl=original", wildcard, hostname)
		r, err := client.Get(tg)
		if err != nil {
			retries -= 1
			if retries == 0 {
				return nil, errors.New(fmt.Sprintf("http request to web.archive.org failed: %s", err.Error()))
			}
			continue
		}
		err = json.NewDecoder(r.Body).Decode(&waybackresp)
		if err != nil {
			retries -= 1
			if retries == 0 {
				return nil, errors.New(fmt.Sprintf("could not decoding response from wayback machine: %s", err.Error()))
			}
			continue
		}
		r.Body.Close()
		break
	}
	first := true
	for _, result := range waybackresp {
		if first {
			// skip first result from wayback machine
			// always is "original"
			first = false
			continue
		}
		found = append(found, result[0])
	}
	return found, nil
}
func getCommonCrawlURLs(domain string) ([]string, error) {
	if CCApi == "" {
		return nil, errors.New("cc api is null")
	}
	var found []string
	wildcard := "*."
	if !IncludeSubs {
		wildcard = ""
	}
	var err error
	var res = &http.Response{}
	retries := MaxRetries
	for retries > 0 {
		res, err = http.Get(
			fmt.Sprintf("%s?url=%s%s/*&output=json", CCApi, wildcard, domain),
		)
		if err != nil {
			retries -= 1
			if retries == 0 {
				return nil, err
			}
			continue
		}
		break
	}
	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)

	for sc.Scan() {
		wrapper := struct {
			URL string `json:"url"`
		}{}
		err = json.Unmarshal([]byte(sc.Text()), &wrapper)

		if err != nil {
			continue
		}

		found = append(found, wrapper.URL)
	}
	return found, nil
}
func getCurrentCC() (string, error) {
	retries := MaxRetries
	var r *http.Response
	wrapper := []struct {
		API string `json:"cdx-api"`
	}{}

	for retries > 0 {
		r, err := client.Get("http://index.commoncrawl.org/collinfo.json")
		if err != nil {
			retries -= 1
			if retries == 0 {
				return "", errors.New(fmt.Sprintf("could not make http request to commoncrawl: %s", err.Error()))
			}
			continue
		}
		err = json.NewDecoder(r.Body).Decode(&wrapper)
		if err != nil {
			retries -= 1
			if retries == 0 {
				return "", errors.New(fmt.Sprintf("could not unmarshal json from commoncrawl: %s", err.Error()))
			}
			continue
		}
	}
	r.Body.Close()
	if len(wrapper) < 1 {
		return "", errors.New("unexpected response from commoncrawl.")
	}
	return wrapper[0].API, nil
}
