package main

import (
	"bufio"
	"github.com/lc/gau/v2/pkg/output"
	"github.com/lc/gau/v2/runner"
	"github.com/lc/gau/v2/runner/flags"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

func main() {
	flag := flags.New()
	cfg, err := flag.ReadInConfig()
	if err != nil {
		if cfg.Verbose {
			log.Warnf("error reading config: %v", err)
		}
	}

	pMap := make(runner.ProvidersMap)
	for _, provider := range cfg.Providers {
		pMap[provider] = cfg.Filters
	}

	config, err := cfg.ProviderConfig()
	if err != nil {
		log.Fatal(err)
	}

	gau := &runner.Runner{}

	if err = gau.Init(config, pMap); err != nil {
		log.Fatal(err)
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

	writeWg := &sync.WaitGroup{}
	writeWg.Add(1)
	if config.JSON {
		go func() {
			defer writeWg.Done()
			output.WriteURLsJSON(out, results, config.Blacklist)
		}()
	} else {
		go func() {
			defer writeWg.Done()
			if err = output.WriteURLs(out, results, config.Blacklist); err != nil {
				log.Fatalf("error writing results: %v\n", err)
			}
		}()
	}

	domains := make(chan string)
	gau.Start(domains, results)

	if len(flags.Args()) > 0 {
		for _, domain := range flags.Args() {
			domains <- domain
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			domains <- sc.Text()
		}

		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	}

	close(domains)

	// wait for providers to fetch URLS
	gau.Wait()

	// close results channel
	close(results)

	// wait for writer to finish output
	writeWg.Wait()
}
