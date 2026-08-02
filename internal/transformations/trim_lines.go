package transformations

type TrimLines struct{}

func (t TrimLines) Name() string {
	return "TrimLines"
}

func (t TrimLines)Apply(data []byte)([]byte, error){
	return data, nil
}