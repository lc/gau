package providers

import (
	"context"

	"github.com/lc/gau/v2/pkg/output"
	"github.com/valyala/fasthttp"
)

const Version = `2.1.2`

// Provider is a generic interface for all archive fetchers
type Provider interface {
	Fetch(ctx context.Context, domain string, results chan output.Result) error
	Name() string
}

type URLScan struct {
	Host   string
	APIKey string
}

type Config struct {
	Threads           uint
	Timeout           uint
	Verbose           bool
	MaxRetries        uint
	IncludeSubdomains bool
	RemoveParameters  bool
	Client            *fasthttp.Client
	Providers         []string
	Blacklist         map[string]struct{}
	Output            string
	JSON              bool
	URLScan           URLScan
	OTX               string
}
