package jsonx

import (
	"encoding/json"
	"os"
)

func ToJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func ToPrettyJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func FromJSON[T any](s string) (T, error) {
	var out T
	err := json.Unmarshal([]byte(s), &out)
	return out, err
}
func WriteJSONFile(path string, v any, pretty bool) error {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
func ReadJSONFile[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}
