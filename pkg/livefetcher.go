package pkg

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("instance not found")

// LiveFetcher defines how instances would be retrieved from a live source.
type LiveFetcher interface {
	Fetch(ctx context.Context) (InstanceMap, error)
}
