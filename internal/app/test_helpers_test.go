package app

import (
	"bytes"
	"encoding/json"
	"testing"
)

func decodeJSONMap(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}
