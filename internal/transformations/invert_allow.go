package transformations

type InvertAllow struct{}

func (i InvertAllow) Name() string {
	return "InvertAllow"
}

func (i InvertAllow) Apply(data []byte) ([]byte, error) {
	return data, nil
}
