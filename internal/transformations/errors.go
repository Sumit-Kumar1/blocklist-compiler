package transformations

import (
	"errors"
	"fmt"
)

var (
	// ErrCancelled reports that the context ended mid-transformation.
	ErrCancelled = errors.New("context cancelled")
	// ErrUnknownTransformation reports a name the factory cannot build.
	ErrUnknownTransformation = errors.New("unknown transformation")
)

func cancelled(name string) error {
	return fmt.Errorf("%s: %w", name, ErrCancelled)
}
