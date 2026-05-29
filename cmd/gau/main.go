package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"

	"github.com/lc/gau/v2/pkg/output"
	"github.com/lc/gau/v2/runner"
	"github.com/lc/gau/v2/runner/flags"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := flags.New().ReadInConfig()
	if err != nil {
		log.Warnf("error reading config: %v", err)
	}

	config, err := cfg.ProviderConfig()
	if err != nil {
		log.Fatal(err)
	}

	gau := new(runner.Runner)

	if err = gau.Init(config, cfg.Providers, cfg.Filters); err != nil {
		log.Warn(err)
	}

	results := make(chan string)

	out := os.Stdout
	// Handle results in background
	if config.Output != "" {
		out, err = os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Fatalf("Could not open output file: %v\n", err)
		}
		defer out.Close()
	}

	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go func(out io.Writer, JSON bool) {
		defer writeWg.Done()
		if JSON {
			output.WriteURLsJSON(out, results, config.Blacklist, config.RemoveParameters)
		} else if err = output.WriteURLs(out, results, config.Blacklist, config.RemoveParameters); err != nil {
			log.Fatalf("error writing results: %v\n", err)
		}
	}(out, config.JSON)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workChan := make(chan runner.Work)
	gau.Start(ctx, workChan, results)
	domains := flags.Args()
	if len(domains) > 0 {
		for _, provider := range gau.Providers {
			for _, domain := range domains {
				workChan <- runner.NewWork(domain, provider)
			}
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			domain := sc.Text()
			for _, provider := range gau.Providers {
				workChan <- runner.NewWork(domain, provider)
			}
		}
		if err := sc.Err(); err != nil {
			log.Fatal(err)
		}
	}
	close(workChan)

	// wait for providers to fetch URLS
	gau.Wait()

	// close results channel
	close(results)

	// wait for writer to finish output
	writeWg.Wait()
}
