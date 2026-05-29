package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/lc/gau/v2/pkg/providers"
	"github.com/lc/gau/v2/pkg/providers/commoncrawl"
	"github.com/lc/gau/v2/pkg/providers/otx"
	"github.com/lc/gau/v2/pkg/providers/urlscan"
	"github.com/lc/gau/v2/pkg/providers/wayback"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	sync.WaitGroup

	Providers []providers.Provider
	threads   uint
	ctx       context.Context
}

// Init initializes the runner
func (r *Runner) Init(c *providers.Config, providers []string, filters providers.Filters) error {
	r.threads = c.Threads
	for _, name := range providers {
		switch name {
		case "urlscan":
			r.Providers = append(r.Providers, urlscan.New(c))
		case "otx":
			r.Providers = append(r.Providers, otx.New(c))
		case "wayback":
			r.Providers = append(r.Providers, wayback.New(c, filters))
		case "commoncrawl":
			cc, err := commoncrawl.New(c, filters)
			if err != nil {
				return fmt.Errorf("error instantiating commoncrawl: %v\n", err)
			}
			r.Providers = append(r.Providers, cc)
		}
	}

	return nil
}

// Starts starts the worker
func (r *Runner) Start(ctx context.Context, workChan chan Work, results chan string) {
	for i := uint(0); i < r.threads; i++ {
		r.Add(1)
		go func() {
			defer r.Done()
			r.worker(ctx, workChan, results)
		}()
	}
}

type Work struct {
	domain   string
	provider providers.Provider
}

func NewWork(domain string, provider providers.Provider) Work {
	return Work{domain, provider}
}

func (w *Work) Do(ctx context.Context, results chan string) error {
	return w.provider.Fetch(ctx, w.domain, results)
}

// worker checks to see if the context is finished and executes the fetching process for each provider
func (r *Runner) worker(ctx context.Context, workChan chan Work, results chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case work, ok := <-workChan:
			if !ok {
				return
			}
			if err := work.Do(ctx, results); err != nil {
				logrus.WithField("provider", work.provider.Name()).Warnf("%s - %v", work.domain, err)
			}
		}
	}
}
