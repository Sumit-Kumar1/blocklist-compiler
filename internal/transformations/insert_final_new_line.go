package transformations

// InsertFinalNewLine ensures there's a final newline
type InsertFinalNewLine struct{}

func (i InsertFinalNewLine) Apply(data []byte) ([]byte, error) {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return append(data, '\n'), nil
	}
	return data, nil
}

func (i InsertFinalNewLine) Name() string {
	return "InsertFinalNewLine"
}
