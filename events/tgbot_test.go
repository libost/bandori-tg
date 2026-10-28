package events

import (
	"encoding/json/jsontext"
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
		{name: "json number", in: jsontext.Value("42"), want: "42"},
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

func TestAllEqualPercent(t *testing.T) {
	tests := []struct {
		name     string
		percents []int
		want     bool
	}{
		{name: "all equal", percents: []int{10, 10, 10}, want: true},
		{name: "not all equal", percents: []int{10, 20, 10}, want: false},
		{name: "single element", percents: []int{10}, want: true},
		{name: "empty slice", percents: []int{}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allEqualPercent(tt.percents); got != tt.want {
				t.Fatalf("allEqualPercent(%v) = %t, want %t", tt.percents, got, tt.want)
			}
		})
	}
}
