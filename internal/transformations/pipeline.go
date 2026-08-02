package transformations

// TransformationPipeline applies transformations in sequence
type TransformationPipeline struct {
	transformations []Transformation
}

// Add adds a transformation to the pipeline
func (p *TransformationPipeline) Add(t Transformation) {
	p.transformations = append(p.transformations, t)
}

// Apply runs all transformations in order, returning first error
func (p *TransformationPipeline) Apply(data []byte) ([]byte, error) {
	for _, t := range p.transformations {
		var err error
		data, err = t.Apply(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}
