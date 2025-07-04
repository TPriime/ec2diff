package pkg

import (
	"context"
	"fmt"
)

var ErrNotFound = fmt.Errorf("instance not found")

type Client interface {
	GetInstance(ctx context.Context, instanceID string) (*Instance, error)
}
