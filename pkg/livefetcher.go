package pkg

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("instance not found")

// PaginatedLiveFetcher defines how instances would be retrieved from a live source.
type PaginatedLiveFetcher interface {
	Fetch(ctx context.Context, onPpageFn func(page int, instances InstanceMap) bool) error
}
