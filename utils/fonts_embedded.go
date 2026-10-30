//go:build !fonts_detached

package utils

import (
	"archive/zip"
	"bytes"
	"embed"
	"fmt"
	"io"
	"strings"
)

//go:embed fonts/*.zip
var fontFS embed.FS

var FontsDetached = false

func GetFontsFile(fontName string) ([]byte, error) {
	fontName = strings.TrimSuffix(fontName, ".ttf")
	fontName += ".zip"
	zipFilePath := "fonts/" + fontName
	fontData, err := fontFS.ReadFile(zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file %s: %w", zipFilePath, err)
	}
	reader, err := zip.NewReader(bytes.NewReader(fontData), int64(len(fontData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader for file %s: %w", zipFilePath, err)
	}
	for _, file := range reader.File {
		if strings.HasSuffix(file.Name, ".ttf") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open font file %s inside zip: %w", file.Name, err)
			}
			defer rc.Close()
			fontBytes, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read font file %s inside zip: %w", file.Name, err)
			}
			return fontBytes, nil
		}
	}
	return nil, fmt.Errorf("font file not found in zip: %s", zipFilePath)
}

func InitFontsFiles() error {
	// No initialization needed for embedded fonts
	// But the complier expects this function to exist, so we provide a no-op implementation.
	return nil
}
