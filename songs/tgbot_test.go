package songs

import (
	"fmt"
	"image/png"
	"os"
	"testing"
)

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
