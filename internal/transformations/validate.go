package transformations

type Validate struct{}

func (v Validate) Name() string {
	return "Validate"
}

func (v Validate) Apply(data []byte) ([]byte, error) {
	return data, nil
}
