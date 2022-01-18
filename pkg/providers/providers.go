package providers

import (
	"context"
	"github.com/valyala/fasthttp"
)

const Version = `2.0.7`

// Provider is a generic interface for all archive fetchers
type Provider interface {
	Fetch(ctx context.Context, domain string, results chan string) error
	Name() string
}

type URLScan struct {
	Host   string
	APIKey string
}

type Config struct {
	Threads           uint
	Verbose           bool
	MaxRetries        uint
	IncludeSubdomains bool
	Client            *fasthttp.Client
	Providers         []string
	Blacklist         map[string]struct{}
	Output            string
	JSON              bool
	URLScan           URLScan
	OTX               string
}
