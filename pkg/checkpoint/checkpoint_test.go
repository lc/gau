package checkpoint

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestNewCheckpoint(t *testing.T) {
	c := New()

	if c.IsCompleted("example.com", "wayback") {
		t.Fatal("new checkpoint should have no completed jobs")
	}
}

func TestMarkCompleted(t *testing.T) {
	c := New()

	c.MarkCompleted("example.com", "wayback")

	if !c.IsCompleted("example.com", "wayback") {
		t.Fatal("job should be marked as completed")
	}

	if c.IsCompleted("example.com", "urlscan") {
		t.Fatal("different provider should not be marked completed")
	}
}

func TestSaveAndLoad(t *testing.T) {
	c := New()
	c.MarkCompleted("example.com", "wayback")

	filename := t.TempDir() + "/checkpoint.json"

	if err := c.Save(filename); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	loaded, err := Load(filename)
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if !loaded.IsCompleted("example.com", "wayback") {
		t.Fatal("loaded checkpoint should contain completed job")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			domain := "example" + string(rune('A'+i)) + ".com"
			c.MarkCompleted(domain, "wayback")
			c.IsCompleted(domain, "wayback")
		}(i)
	}

	wg.Wait()
}
func TestProgress(t *testing.T) {
	cp := New()

	// New checkpoint should start at page 0.
	if got := cp.GetProgress("example.com", "wayback"); got != 0 {
		t.Fatalf("expected progress 0, got %d", got)
	}

	cp.SetProgress("example.com", "wayback", 5)

	if got := cp.GetProgress("example.com", "wayback"); got != 5 {
		t.Fatalf("expected progress 5, got %d", got)
	}
}

func TestProgressSaveAndLoad(t *testing.T) {
	cp := New()
	cp.SetProgress("example.com", "wayback", 12)

	filename := filepath.Join(t.TempDir(), "checkpoint.json")

	if err := cp.Save(filename); err != nil {
		t.Fatalf("failed to save checkpoint: %v", err)
	}

	loaded, err := Load(filename)
	if err != nil {
		t.Fatalf("failed to load checkpoint: %v", err)
	}

	if got := loaded.GetProgress("example.com", "wayback"); got != 12 {
		t.Fatalf("expected progress 12 after loading, got %d", got)
	}
}
