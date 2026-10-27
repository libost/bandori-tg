package songs

import (
	"fmt"
	"testing"

	"github.com/libost/gotgbot/v2"
)

func TestIsSpecifiedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "HTTP URL content error",
			err: &gotgbot.TelegramError{
				Method:      "sendRichMessage",
				Code:        400,
				Description: "Bad Request: failed to get HTTP URL content",
			},
			want: true,
		},
		{
			name: "wrapped HTTP URL content error",
			err: fmt.Errorf("send song details: %w", &gotgbot.TelegramError{
				Method:      "sendRichMessage",
				Code:        400,
				Description: "Bad Request: failed to get HTTP URL content",
			}),
			want: true,
		},
		{
			name: "different method",
			err: &gotgbot.TelegramError{
				Method:      "sendMessage",
				Code:        400,
				Description: "Bad Request: failed to get HTTP URL content",
			},
			want: false,
		},
		{
			name: "different error",
			err: &gotgbot.TelegramError{
				Method:      "sendRichMessage",
				Code:        400,
				Description: "Bad Request: chat not found",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSpecifiedError(tt.err); got != tt.want {
				t.Fatalf("isSpecifiedError(%v) = %t, want %t", tt.err, got, tt.want)
			}
		})
	}
}
