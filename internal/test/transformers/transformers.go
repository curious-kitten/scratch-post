package transformers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

// ToReadCloser transforms a truct to a ReadCloser object
func ToReadCloser(item interface{}) io.ReadCloser {
	value, _ := json.Marshal(item)
	reader := bytes.NewReader(value)
	return ioutil.NopCloser(reader)
}
