package transformations

type InsertFinalNewLine struct{}

func (i InsertFinalNewLine) Name() string {
	return "InsertFinalNewLine"
}

func (i InsertFinalNewLine) Apply(data []byte) ([]byte, error) {
	return data, nil
}
