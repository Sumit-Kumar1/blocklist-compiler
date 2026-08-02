package transformations

type RemoveEmptyLines struct{}

func (r RemoveEmptyLines) Name() string {
	return "RemoveEmptyLines"
}

func (r RemoveEmptyLines) Apply(data []byte) ([]byte, error) {
	return data, nil
}
