package main

import (
	"bufio"
	"crypto/tls"
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

	"github.com/lc/gau/output"
	"github.com/lc/gau/providers"
)

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
			output.WriteURLsJSON(results, out, config.Blacklist)
			writewg.Done()
		}()
	} else {
		go func() {
			output.WriteURLs(results, out, config.Blacklist)
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
	proxy := flag.String("p", "", "HTTP proxy to use")
	output := flag.String("o", "", "filename to write results to")
	jsonOut := flag.Bool("json", false, "write output as json")
	blacklist := flag.String("b","","extensions to skip, ex: ttf,woff,svg,png,jpg")
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
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	if *proxy != "" {
		if p, err := url.Parse(*proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	extensions := strings.Split(*blacklist,",")
	extMap := make(map[string]struct{})
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = struct{}{}
	}
	config := providers.Config{
		Verbose:           *verbose,
		MaxRetries:        *maxRetries,
		IncludeSubdomains: *includeSubs,
		Output:            *output,
		JSON:              *jsonOut,
		Blacklist: 			extMap,
		Client: &http.Client{
			Timeout:   time.Second * 15,
			Transport: tr,
		},
		Providers: strings.Split(*useProviders, ","),
	}
	run(&config, domains)
}
