package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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

var IncludeSubs bool

var client = &http.Client{
	Timeout: time.Second * 15,
}

func main() {
	var domains []string
	flag.BoolVar(&IncludeSubs, "subs", false, "include subdomains of target domain")
	flag.Parse()
	if flag.NArg() > 0 {
		domains = []string{flag.Arg(0)}
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			domains = append(domains, s.Text())
		}
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
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
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
	for {
		r, err := client.Get(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/hostname/%s/url_list?limit=50&page=%d", hostname, page))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("http request to OTX failed: %s", err.Error()))
		}
		defer r.Body.Close()
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("error reading body from alienvault: %s", err.Error()))
		}
		o := &OTXResult{}
		err = json.Unmarshal(bytes, o)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("could not decode json response from alienvault: %s", err.Error()))
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
	var found []string
	tg := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey&fl=original", wildcard, hostname)
	r, err := client.Get(tg)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("http request to web.archive.org failed: %s", err.Error()))
	}
	defer r.Body.Close()
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading body: %s", err.Error()))
	}
	err = json.Unmarshal(resp, &waybackresp)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decoding response from wayback machine: %s", err.Error()))
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
	var found []string
	wildcard := "*."
	if !IncludeSubs {
		wildcard = ""
	}
	currentApi, err := getCurrentCC()
	if err != nil {
		return nil, fmt.Errorf("error getting current commoncrawl url: %v", err)
	}
	res, err := http.Get(
		fmt.Sprintf("%s?url=%s%s/*&output=json", currentApi, wildcard, domain),
	)
	if err != nil {
		return nil, err
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
	r, err := client.Get("http://index.commoncrawl.org/collinfo.json")
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	wrapper := []struct {
		API string `json:"cdx-api"`
	}{}
	err = json.Unmarshal(resp, &wrapper)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal json from CC: %s", err.Error())
	}
	if len(wrapper) < 1 {
		return "", errors.New("unexpected response from commoncrawl.")
	}
	return wrapper[0].API, nil
}
