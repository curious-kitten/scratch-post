package keys

import (
	"bytes"
	"io"
)

type Retriever struct {
	Item io.ReadCloser
}

func (r *Retriever) GetOne() ([]byte, error) {
	defer r.Item.Close()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Item)
	if err != nil {
		return []byte{}, nil
	}
	return buf.Bytes(), nil
}
