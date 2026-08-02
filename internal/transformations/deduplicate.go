package transformations

type Deduplicate struct{}

func (dd Deduplicate) Name() string {
	return "Deduplicate"
}

func (dd Deduplicate) Apply(data []byte) ([]byte, error) {
	return data, nil
}
