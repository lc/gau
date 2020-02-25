package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

var IncludeSubs bool

var client = &http.Client{
	Timeout: time.Second * 20,
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
	tmp := fmt.Sprintf("%s", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))
	_ = tmp
	fetchers := []fetch{getWaybackUrls, getCommonCrawlURLs, getOtxUrls}
	for _, fn := range fetchers {
		found, err := fn(domain)
		if err != nil {
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
			log.Fatalf("Error: %v", err)
		}
		defer r.Body.Close()
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		o := &OTXResult{}
		err = json.Unmarshal(bytes, o)
		if err != nil {
			log.Fatalf("Could not decode json: %s\n", err)
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
		log.Printf("Error in http request: %v\n", err)
		return nil, err
	}
	defer r.Body.Close()
	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		return nil, err
	}
	err = json.Unmarshal(resp, &waybackresp)
	if err != nil {
		log.Printf("Error unmarshalling response: %v\n", err)
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
	res, err := http.Get(
		fmt.Sprintf("http://index.commoncrawl.org/CC-MAIN-2019-43-index?url=%s%s/*&output=json", wildcard, domain),
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
