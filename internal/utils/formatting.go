package utils

import (
	"encoding/json"
)

// MakeJSONStringUniform takes a JSON string and returns it as a JSON string with uniform formatting
func MakeJSONStringUniform(input json.RawMessage) (string, error) {
	var settingsJSON map[string]interface{}
	err := json.Unmarshal(input, &settingsJSON)
	if err != nil {
		return "", err
	}

	output, err := json.Marshal(settingsJSON)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
