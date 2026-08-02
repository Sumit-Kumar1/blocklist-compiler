package processor

type Transformation interface {
	Apply(data []byte) ([]byte, error)
	Name() string
}
