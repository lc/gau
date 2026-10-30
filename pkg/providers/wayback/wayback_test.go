package wayback

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lc/gau/v2/pkg/checkpoint"
	"github.com/lc/gau/v2/pkg/providers"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchResumesFromSavedPage(t *testing.T) {
	cp := checkpoint.New()
	cp.SetProgress("example.com", Name, 5)

	var requestedPages []string

	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			page := req.URL.Query().Get("page")
			requestedPages = append(requestedPages, page)

			body := `[["original"],["https://example.com/test"]]`

			if page == "6" {
				body = `[]`
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}

	config := &providers.Config{
		HTTPClient: httpClient,
		Checkpoint: cp,
	}

	client := New(config, providers.Filters{})
	results := make(chan string, 10)

	err := client.Fetch(context.Background(), "example.com", results)
	if err != nil {
		t.Fatalf("Fetch returned an unexpected error: %v", err)
	}

	if len(requestedPages) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requestedPages))
	}

	if requestedPages[0] != "5" {
		t.Errorf("expected first request to page 5, got page %s", requestedPages[0])
	}

	if requestedPages[1] != "6" {
		t.Errorf("expected second request to page 6, got page %s", requestedPages[1])
	}

	select {
	case result := <-results:
		if result != "https://example.com/test" {
			t.Errorf("unexpected result: %s", result)
		}
	default:
		t.Error("expected a URL result, but received none")
	}

	if got := cp.GetProgress("example.com", Name); got != 6 {
		t.Errorf("expected saved progress to be 6, got %d", got)
	}
}

func TestFetchStopsWhenContextIsCancelled(t *testing.T) {
	httpClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body := `[["original"],["https://example.com/test"]]`

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}),
	}

	config := &providers.Config{
		HTTPClient: httpClient,
		Checkpoint: checkpoint.New(),
	}

	client := New(config, providers.Filters{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := make(chan string)

	err := client.Fetch(ctx, "example.com", results)

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
