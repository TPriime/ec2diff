package pkg

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("instance not found")

type LiveFetcher interface {
	GetInstance(ctx context.Context, instanceID string) (*Instance, error)
}
