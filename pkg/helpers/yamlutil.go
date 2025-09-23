package helpers

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ToYAML marshals a value to YAML string.
func ToYAML(v any) (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromYAML unmarshals YAML string into a typed value.
func FromYAML[T any](s string) (T, error) {
	var out T
	err := yaml.Unmarshal([]byte(s), &out)
	return out, err
}

// WriteYAMLFile writes the given value to a YAML file with 0644 perms.
func WriteYAMLFile(path string, v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ReadYAMLFile reads a YAML file into a typed value.
func ReadYAMLFile[T any](path string) (T, error) {
	var out T
	b, err := os.ReadFile(path)
	if err != nil {
		return out, err
	}
	err = yaml.Unmarshal(b, &out)
	return out, err
}
