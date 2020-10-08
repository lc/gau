package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"crypto/tls"
	"math/rand"

	"github.com/lc/gau/output"
	"github.com/lc/gau/providers"
)


func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))

	pick := payload[randomIndex]

	return pick
}

func run(config *providers.Config, domains []string) {
	var providerList []providers.Provider

	for _, toUse := range config.Providers {
		switch toUse {
		case "wayback":
			wayback := providers.NewWaybackProvider(config)
			providerList = append(providerList, wayback)
		case "otx":
			otx := providers.NewOTXProvider(config)
			providerList = append(providerList, otx)
		case "commoncrawl":
			common, err := providers.NewCommonProvider(config)
			if err == nil {
				providerList = append(providerList, common)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: %s is not a valid provider.\n", toUse)
		}
	}

	results := make(chan string)
	var out io.Writer
	// Handle results in background
	if config.Output == "" {
		out = os.Stdout
	} else {
		ofp, err := os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Could not open output file: %v\n", err)
		}
		defer ofp.Close()
		out = ofp
	}
	writewg := &sync.WaitGroup{}
	writewg.Add(1)
	if config.JSON {
		go func() {
			output.WriteURLsJSON(results, out)
			writewg.Done()
		}()
	} else {
		go func() {
			output.WriteURLs(results, out)
			writewg.Done()
		}()
	}
	exitStatus := 0
	for _, domain := range domains {
		// Run all providers in parallel
		wg := &sync.WaitGroup{}
		wg.Add(len(providerList))

		for _, provider := range providerList {
			go func(provider providers.Provider) {
				defer wg.Done()
				if err := provider.Fetch(domain, results); err != nil {
					if config.Verbose {
						_, _ = fmt.Fprintln(os.Stderr, err)
					}
				}
			}(provider)
		}
		// Wait for providers to finish their tasks
		wg.Wait()
	}
	close(results)
	// Wait for writer to finish
	writewg.Wait()
	os.Exit(exitStatus)
}

func main() {
	var domains []string
	verbose := flag.Bool("v", false, "enable verbose mode")
	includeSubs := flag.Bool("subs", false, "include subdomains of target domain")
	maxRetries := flag.Uint("retries", 5, "amount of retries for http client")
	useProviders := flag.String("providers", "wayback,otx,commoncrawl", "providers to fetch urls for")
	version := flag.Bool("version", false, "show gau version")
	proxy := flag.String("p", "", "use proxy")
	output := flag.String("o", "", "filename to write results to")
	jsonOut := flag.Bool("json", false, "write output as json")
	randomAgent := flag.Bool("random-agent", false, "use random user-agent")
	flag.Parse()

	if *version {
		fmt.Printf("gau version: %s\n", providers.Version)
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		domains = flag.Args()
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			domains = append(domains, s.Text())
		}
	}
	
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	if *proxy != "" {
		if p, err := url.Parse(*proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	config := providers.Config{
		Verbose:           *verbose,
		RandomAgent:	   *randomAgent,
		MaxRetries:        *maxRetries,
		IncludeSubdomains: *includeSubs,
		Output:            *output,
		JSON:              *jsonOut,
		Client: &http.Client{
			Timeout: time.Second * 15,
			Transport: tr,
		},
		Providers: strings.Split(*useProviders, ","),
	}
	run(&config, domains)
}
