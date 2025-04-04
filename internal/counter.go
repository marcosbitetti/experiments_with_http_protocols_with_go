package internal

import (
	"sync"
)

// Counter is a thread-safe counter implementation
type Counter struct {
	mu    sync.Mutex
	count int64
}

// NewCounter creates a new Counter instance
func NewCounter() *Counter {
	return &Counter{
		count: 0,
	}
}

// GetCount returns the current count value in a thread-safe manner
func (c *Counter) GetCountAndIncrement() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.count
	c.count++
	return current
}

// ResetCount resets the counter to zero and returns the previous value
func (c *Counter) ResetCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count = 1
}
