package transformations

type Compress struct{}

func (c Compress) Name() string {
	return "Compress"
}

func (c Compress) Apply(data []byte) ([]byte, error) {
	return data, nil
}
