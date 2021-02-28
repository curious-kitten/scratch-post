package keys

import (
	"bytes"
	"io"
)

// Retriever is used to get the key used to generate the JWT from a file.
// This is done to help with secret rotation
type Retriever struct {
	Item io.ReadCloser
}

// GetOne returns the contents of the file containing the secret
func (r *Retriever) GetOne() ([]byte, error) {
	defer r.Item.Close()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Item)
	if err != nil {
		return []byte{}, nil
	}
	return buf.Bytes(), nil
}
