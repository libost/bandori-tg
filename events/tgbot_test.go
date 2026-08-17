package events

import (
	"encoding/json"
	"testing"
)

func TestNormalizeCharacterID(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "int", in: 42, want: "42"},
		{name: "int64", in: int64(42), want: "42"},
		{name: "float64", in: float64(42), want: "42"},
		{name: "json number", in: json.Number("42"), want: "42"},
		{name: "string", in: "42", want: "42"},
		{name: "nil", in: nil, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeCharacterID(tt.in); got != tt.want {
				t.Fatalf("normalizeCharacterID(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
