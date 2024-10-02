package utils

import (
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

// Serializes JSON string into provided interface
func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)

	return e.Encode(i)
}