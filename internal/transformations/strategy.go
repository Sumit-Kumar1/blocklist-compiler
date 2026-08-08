package transformations

import (
	"context"
)

// Transformation defines the contract for all blocklist transformations
type Transformation interface {
	Apply(ctx context.Context, data []byte) ([]byte, error)
	Name() string
}
