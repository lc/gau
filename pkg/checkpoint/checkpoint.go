package checkpoint

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Checkpoint stores completed domain-provider jobs.
type Checkpoint struct {
	mu        sync.Mutex
	Completed map[string]bool `json:"completed"`
	Progress  map[string]uint `json:"progress"`
}

// New creates an empty checkpoint.
func New() *Checkpoint {
	return &Checkpoint{
		Completed: make(map[string]bool),
		Progress:  make(map[string]uint),
	}
}

// jobKey generates a unique key for a domain-provider combination.
func jobKey(domain, provider string) string {
	return domain + ":" + provider
}

// IsCompleted checks whether a job has already been completed.
func (c *Checkpoint) IsCompleted(domain, provider string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Completed[jobKey(domain, provider)]
}

// MarkCompleted marks a domain-provider job as completed.
func (c *Checkpoint) MarkCompleted(domain, provider string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Completed == nil {
		c.Completed = make(map[string]bool)
	}

	c.Completed[jobKey(domain, provider)] = true
}

// GetProgress returns the next page to fetch for a domain-provider job.
func (c *Checkpoint) GetProgress(domain, provider string) uint {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Progress[jobKey(domain, provider)]
}

// SetProgress stores the next page to fetch for a domain-provider job.
func (c *Checkpoint) SetProgress(domain, provider string, nextPage uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Progress == nil {
		c.Progress = make(map[string]uint)
	}

	c.Progress[jobKey(domain, provider)] = nextPage
}

// Save writes checkpoint data using atomic file replacement.
func (c *Checkpoint) Save(filename string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)

	tmp, err := os.CreateTemp(dir, ".checkpoint-*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, filename)
}

// Load reads an existing checkpoint file.
func Load(filename string) (*Checkpoint, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := New()

	if err := json.Unmarshal(data, c); err != nil {
		return nil, err
	}

	if c.Completed == nil {
		c.Completed = make(map[string]bool)
	}
	if c.Progress == nil {
		c.Progress = make(map[string]uint)
	}

	return c, nil
}
