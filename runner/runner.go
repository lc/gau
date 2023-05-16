package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/lc/gau/v2/pkg/output"
	"github.com/lc/gau/v2/pkg/providers"
	"github.com/lc/gau/v2/pkg/providers/commoncrawl"
	"github.com/lc/gau/v2/pkg/providers/otx"
	"github.com/lc/gau/v2/pkg/providers/urlscan"
	"github.com/lc/gau/v2/pkg/providers/wayback"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	providers []providers.Provider
	wg        sync.WaitGroup

	config     *providers.Config
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type ProvidersMap map[string]providers.Filters

// Init initializes the runner
func (r *Runner) Init(c *providers.Config, providerMap ProvidersMap) error {
	r.config = c
	r.ctx, r.cancelFunc = context.WithCancel(context.Background())

	for name, filters := range providerMap {
		switch name {
		case urlscan.Name:
			r.providers = append(r.providers, urlscan.New(c))
		case otx.Name:
			o := otx.New(c)
			r.providers = append(r.providers, o)
		case wayback.Name:
			r.providers = append(r.providers, wayback.New(c, filters))
		case commoncrawl.Name:
			cc, err := commoncrawl.New(c, filters)
			if err != nil {
				return fmt.Errorf("error instantiating commoncrawl: %v\n", err)
			}
			r.providers = append(r.providers, cc)
		}
	}

	return nil
}

// Starts starts the worker
func (r *Runner) Start(domains chan string, results chan output.Result) {
	for i := uint(0); i < r.config.Threads; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			r.worker(r.ctx, domains, results)
		}()
	}
}

// Wait waits for the providers to finish fetching
func (r *Runner) Wait() {
	r.wg.Wait()
}

// worker checks to see if the context is finished and executes the fetching process for each provider
func (r *Runner) worker(ctx context.Context, domains chan string, results chan output.Result) {
work:
	for {
		select {
		case <-ctx.Done():
			break work
		case domain, ok := <-domains:
			if ok {
				var wg sync.WaitGroup
				for _, p := range r.providers {
					wg.Add(1)
					go func(p providers.Provider) {
						defer wg.Done()
						if err := p.Fetch(ctx, domain, results); err != nil && r.config.Verbose {
							logrus.WithField("provider", p.Name()).Warnf("%s - %v", domain, err)
						}
					}(p)
				}
				wg.Wait()
			}
			if !ok {
				break work
			}
		}
	}
}
