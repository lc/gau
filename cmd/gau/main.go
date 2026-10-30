package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/lc/gau/v2/pkg/checkpoint"
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

	// Initialize or restore checkpoint.
	checkpointPath := cfg.CheckpointFile

	var cp *checkpoint.Checkpoint

	if cfg.ResumeFile != "" {
		cp, err = checkpoint.Load(cfg.ResumeFile)
		if err != nil {
			log.Fatalf("Could not load checkpoint: %v", err)
		}

		// Continue saving to the resume file unless another
		// checkpoint output file was specified.
		if checkpointPath == "" {
			checkpointPath = cfg.ResumeFile
		}
	} else {
		cp = checkpoint.New()

		// Avoid accidentally overwriting an existing checkpoint.
		if checkpointPath != "" {
			if _, statErr := os.Stat(checkpointPath); statErr == nil {
				log.Fatalf("Checkpoint already exists. Use --resume %s", checkpointPath)
			} else if !os.IsNotExist(statErr) {
				log.Fatalf("Could not access checkpoint file: %v", statErr)
			}
		}
	}

	gau.Checkpoint = cp
	gau.CheckpointFile = checkpointPath
	config.Checkpoint = cp
	config.CheckpointFile = checkpointPath
	if checkpointPath != "" {
		if err := cp.Save(checkpointPath); err != nil {
			log.Fatalf("Could not create checkpoint file: %v", err)
		}
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
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()
	go func() {
		<-ctx.Done()
		log.Warn("Interrupt signal received! Shutting down...")
	}()
	workChan := make(chan runner.Work)
	gau.Start(ctx, workChan, results)
	scheduleWork := func(domain string) bool {
		for _, provider := range gau.Providers {
			select {
			case workChan <- runner.NewWork(domain, provider):
			case <-ctx.Done():
				return false
			}
		}
		return true
	}

	domains := flags.Args()

	if len(domains) > 0 {
		for _, domain := range domains {
			if !scheduleWork(domain) {
				break
			}
		}
	} else {
		sc := bufio.NewScanner(os.Stdin)

		for sc.Scan() {
			if !scheduleWork(sc.Text()) {
				break
			}
		}

		if err := sc.Err(); err != nil && ctx.Err() == nil {
			log.Errorf("Error reading stdin: %v", err)
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
