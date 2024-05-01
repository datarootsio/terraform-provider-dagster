package testutils

import (
	"encoding/json"
)

func UnmarshalJSONOrPanic(input json.RawMessage) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(input, &result)
	if err != nil {
		panic(err)
	}
	return result
}

func MarshalJSONOrPanic(input map[string]interface{}) json.RawMessage {
	output, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return output
}
