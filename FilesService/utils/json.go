package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// Deserializes the object from the reader
func FromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	
	return d.Decode(i)
}

// Deserializes the object from a JSON string
func FromJSONString(i interface{}, jsonStr string) error {
	return FromJSON(i, strings.NewReader(jsonStr))
}

// Serializes provided interface into JSON and writes to writer
func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	e.SetEscapeHTML(false)

	return e.Encode(i)
}

// Serializes interface and returns JSON string
func ToJSONString(i interface{}) (string, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(i)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}