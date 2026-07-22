package semanticcache

import (
	"context"
	"time"
)

type VectorSearch interface {
	FindSimilar(ctx context.Context, prompt string) (float64, string, error)
}

type Cache struct {
	search       VectorSearch
	threshold    float64
	minPromptLen int
	timeout      time.Duration
}

func New(search VectorSearch, threshold float64, minPromptLen int) *Cache {
	return &Cache{
		search:       search,
		threshold:    threshold,
		minPromptLen: minPromptLen,
		timeout:      50 * time.Millisecond,
	}
}

func (c *Cache) Lookup(ctx context.Context, prompt string) (string, bool) {
	if len(prompt) < c.minPromptLen {
		return "", false
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	sim, resp, err := c.search.FindSimilar(ctxTimeout, prompt)
	if err != nil {
		return "", false
	}

	if sim >= c.threshold {
		return resp, true
	}

	return "", false
}
