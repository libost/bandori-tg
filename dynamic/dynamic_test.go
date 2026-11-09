package dynamic

import "testing"

func TestShouldCollectSong(t *testing.T) {
	tests := []struct {
		name        string
		publishedAt []string
		idx         int
		want        bool
	}{
		{
			name:        "target is the largest value",
			publishedAt: []string{"100", "300", "200"},
			idx:         1,
			want:        true,
		},
		{
			name:        "target is not the largest value",
			publishedAt: []string{"300", "100", "200"},
			idx:         1,
			want:        false,
		},
		{
			name:        "target is the only valid value",
			publishedAt: []string{"", "100", ""},
			idx:         1,
			want:        true,
		},
		{
			name:        "empty target",
			publishedAt: []string{"100", "", "200"},
			idx:         1,
			want:        false,
		},
		{
			name:        "invalid values are ignored",
			publishedAt: []string{"invalid", "100", ""},
			idx:         1,
			want:        true,
		},
		{
			name:        "target index is out of range",
			publishedAt: []string{"100"},
			idx:         1,
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCollectSong(tt.publishedAt, tt.idx); got != tt.want {
				t.Fatalf("shouldCollectSong(%v, %d) = %v, want %v", tt.publishedAt, tt.idx, got, tt.want)
			}
		})
	}
}
