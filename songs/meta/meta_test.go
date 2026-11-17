package meta

import "testing"

func Test_timeCalc(t *testing.T) {
	tests := []struct {
		name   string
		length float64
		want   string
	}{
		{
			name:   "length < 60",
			length: 45.034,
			want:   "0:45.0",
		},
		{
			name:   "length = 60",
			length: 60.022,
			want:   "1:00.0",
		},
		{
			name:   "length > 60",
			length: 120.565,
			want:   "2:00.6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeCalc(tt.length)
			if got != tt.want {
				t.Errorf("timeCalc() = %v, want %v", got, tt.want)
			}
		})
	}
}
