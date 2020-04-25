package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lc/gau/providers"
)

// printResults just received fetched URLs and print them.
func printResults(results <-chan string) {
	for result := range results {
		fmt.Println(result)
	}
}

func run(config *providers.Config, domains []string) {
	providerList := []providers.Provider{
		providers.Provider(&providers.WaybackProvider{Config: config}),
		providers.Provider(&providers.OTXProvider{Config: config}),
		providers.Provider(&providers.CommonProvider{Config: config}),
	}

	results := make(chan string)
	defer close(results)

	// Handle results in background
	go printResults(results)

	for _, domain := range domains {
		// Run all providers in parallel
		wg := sync.WaitGroup{}
		wg.Add(len(providerList))

		for _, provider := range providerList {
			go func(provider providers.Provider) {
				defer wg.Done()

				if err := provider.Fetch(domain, results); err != nil {
					_, _ = fmt.Fprintln(os.Stderr, "An error occurred:", err)
				}
			}(provider)
		}

		// Wait for providers to finish their tasks
		wg.Wait()
	}
}

func main() {
	var domains []string

	includeSubs := flag.Bool("subs", false, "include subdomains of target domain")
	maxRetries := flag.Uint("retries", 5, "amount of retries for http client")

	flag.Parse()

	if flag.NArg() > 0 {
		domains = flag.Args()
	} else {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			domains = append(domains, s.Text())
		}
	}

	config := providers.Config{
		MaxRetries:        *maxRetries,
		IncludeSubdomains: *includeSubs,
		Client: &http.Client{
			Timeout: time.Second * 15,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}

	run(&config, domains)
}
