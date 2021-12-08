// Package jsonutil uses json-iterator to get better performance than standard encoding/json package.
// And it's 100% compatible with encoding/json.
package jsonutil

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var defaultStrategy = jsoniter.ConfigCompatibleWithStandardLibrary

func Marshal(v interface{}) ([]byte, error) {
	return defaultStrategy.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return defaultStrategy.MarshalIndent(v, prefix, indent)
}

func MarshalToString(v interface{}) (string, error) {
	return defaultStrategy.MarshalToString(v)
}

func UnmarshalFromString(data string, v interface{}) error {
	return defaultStrategy.UnmarshalFromString(data, v)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return jsoniter.NewEncoder(w)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return jsoniter.NewDecoder(r)
}
