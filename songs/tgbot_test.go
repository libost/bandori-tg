package songs

import (
	"fmt"
	"image/png"
	"os"
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

func TestGenerateDifficultyImage(t *testing.T) {
	testDifficulties := []difficultyData{
		{
			level: []int{5, 10, 15, 20, 25}, // For songs with an extra difficulty "Special"
		},
		{
			level: []int{5, 10, 15, 20}, // For ordinary songs with 4 difficulties
		},
	}
	for i, testDifficulty := range testDifficulties {
		t.Run(fmt.Sprintf("test difficulty image %d", i), func(t *testing.T) {
			img, err := generateDifficultyImg(testDifficulty)
			if err != nil {
				t.Fatalf("generateDifficultyImg() error = %v", err)
			}
			if img == nil {
				t.Fatal("generateDifficultyImg() returned nil image")
			}
			os.MkdirAll("test_output", os.ModePerm)
			outFile, err := os.Create(fmt.Sprintf("test_output/difficulty_image_%d.png", i))
			if err != nil {
				t.Fatalf("Failed to create output file: %v", err)
			}
			defer outFile.Close()
			err = png.Encode(outFile, img)
			if err != nil {
				t.Fatalf("Failed to encode image to PNG: %v", err)
			}
		})
	}
	t.Logf("Difficulty image generated and saved to test_output/difficulty_image.png")
	t.Logf("To ensure the image generated correctly, human inspection is required. Please check the output file.")
}
