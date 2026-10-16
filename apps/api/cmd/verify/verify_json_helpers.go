package main

import "encoding/json"

func jsonUnmarshal(raw []byte, v any) error { return json.Unmarshal(raw, v) }

func jsonMarshalIndent(v any) ([]byte, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}
