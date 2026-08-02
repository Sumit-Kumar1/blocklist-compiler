package transformations

import "errors"

type TransformationFactory struct{}

func (f *TransformationFactory) Create(name string) (Transformation, error) {
	switch name {
	case "RemoveComments":
		return &RemoveComments{}, nil
	case "Deduplicate":
		return &Deduplicate{}, nil
	case "Validate":
		return &Validate{}, nil
	case "Compress":
		return &Compress{}, nil
	case "TrimLines":
		return &TrimLines{}, nil
	case "RemoveEmptyLines":
		return &RemoveEmptyLines{}, nil
	case "InsertFinalNewLine":
		return &InsertFinalNewLine{}, nil
	case "InvertAllow":
		return &InvertAllow{}, nil
	default:
		return nil, errors.New("unknown transformation")
	}
}
