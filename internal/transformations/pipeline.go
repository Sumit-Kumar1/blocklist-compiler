package transformations

import (
	"context"
	"fmt"
)

// TransformationPipeline applies transformations in sequence
type TransformationPipeline struct {
	Transformations []Transformation
}

// Add adds a transformation to the pipeline
func (p *TransformationPipeline) Add(t Transformation) {
	p.Transformations = append(p.Transformations, t)
}

// Apply runs all transformations in order, returning first error
func (p *TransformationPipeline) Apply(ctx context.Context, data []byte) ([]byte, error) {
	for _, t := range p.Transformations {
		var err error

		data, err = t.Apply(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("transformation %q failed: %w", t.Name(), err)
		}
	}

	return data, nil
}
