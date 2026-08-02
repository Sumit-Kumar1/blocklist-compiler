package transformations

// TransformationPipeline applies transformations in sequence
type TransformationPipeline struct {
	transformations []Transformation
}

func NewTransformationPipeline() *TransformationPipeline {
	return &TransformationPipeline{transformations: make([]Transformation, 0)}
}

// Add adds a transformation to the pipeline
func (p *TransformationPipeline) Add(t Transformation) {
	p.transformations = append(p.transformations, t)
}

// Apply runs all transformations in order
func (p *TransformationPipeline) Apply(data []byte) ([]byte, error) {
	var err error

	for _, t := range p.transformations {
		data, err = t.Apply(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
