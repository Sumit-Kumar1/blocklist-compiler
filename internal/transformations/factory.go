package transformations

import "fmt"

// constructors maps a config name to the transformation it builds
var constructors = map[string]func() Transformation{
	"RemoveComments":     func() Transformation { return &RemoveComments{} },
	"Deduplicate":        func() Transformation { return &Deduplicate{} },
	"Validate":           func() Transformation { return &Validate{} },
	"Compress":           func() Transformation { return &Compress{} },
	"TrimLines":          func() Transformation { return &TrimLines{} },
	"RemoveEmptyLines":   func() Transformation { return &RemoveEmptyLines{} },
	"InsertFinalNewLine": func() Transformation { return &InsertFinalNewLine{} },
	"InvertAllow":        func() Transformation { return &InvertAllow{} },
}

// TransformationFactory creates transformation instances by name
type TransformationFactory struct{}

func (f *TransformationFactory) Create(name string) (Transformation, error) {
	newTransformation, ok := constructors[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownTransformation, name)
	}

	return newTransformation(), nil
}
