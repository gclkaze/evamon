package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
)

func GetRandomString() string {
	return uuid.New().String()
}

func GetRandomStringWithLength(n int) string {
	s := strings.ReplaceAll(uuid.NewString(), "-", "")
	if n >= len(s) {
		return s
	}
	return s[:n]
}
func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
func NormalizeSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ============================================================
// Strict JSON helpers
// ============================================================

// decodeStrict decodes JSON and fails on unknown fields.

func DecodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber()

	if err := dec.Decode(v); err != nil {
		return err
	}
	// Ensure no trailing junk
	if dec.More() {
		return errors.New("unexpected trailing JSON content")
	}
	return nil
}
