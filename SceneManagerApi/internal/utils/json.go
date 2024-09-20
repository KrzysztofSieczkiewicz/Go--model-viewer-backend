package utils

import (
	"encoding/json"
	"io"
)

// Deserializes the object from the JSON string
func FromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	
	return d.Decode(i)
}

// Serializes JSON string into provided interface
func ToJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)

	return e.Encode(i)
}