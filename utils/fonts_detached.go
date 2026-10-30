//go:build fonts_detached

package utils

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var FontsDetached = true

const FontsDir = C.BaseDir + "fonts/"

const (
	notoSansJPFontUrl = "https://raw.githubusercontent.com/libost/bandori-tg/main/utils/fonts/" + notoSansJPFontName
	notoSansKRFontUrl = "https://raw.githubusercontent.com/libost/bandori-tg/main/utils/fonts/" + notoSansKRFontName
	notoSansSCFontUrl = "https://raw.githubusercontent.com/libost/bandori-tg/main/utils/fonts/" + notoSansSCFontName
	notoSansTCFontUrl = "https://raw.githubusercontent.com/libost/bandori-tg/main/utils/fonts/" + notoSansTCFontName
	notoSansFontUrl   = "https://raw.githubusercontent.com/libost/bandori-tg/main/utils/fonts/" + notoSansFontName
)

const (
	notoSansJPFontName = "NotoSansJP-Regular.zip"
	notoSansKRFontName = "NotoSansKR-Regular.zip"
	notoSansSCFontName = "NotoSansSC-Regular.zip"
	notoSansTCFontName = "NotoSansTC-Regular.zip"
	notoSansFontName   = "NotoSans-Regular.zip"
)

const (
	notoSansFontSHA256   = "D480D22C09360283D60BDC74B58FB2F1A911247DC5A43F48491A5C1C375E1FB4"
	notoSansJPFontSHA256 = "97A597F81C36E308D07D5F0AFE954BD6F2C234382B77B4ABBE6A7CEF8E134200"
	notoSansKRFontSHA256 = "E0503A905624C20D56A4B0CF8B303CB7BFE96D09A35CC19490962E59264FFF96"
	notoSansSCFontSHA256 = "D432D43F07AB7F8AEA6E2D04449D0D1B7ED95B6CEF3390DCD4EA27BA1F9D77EA"
	notoSansTCFontSHA256 = "1BACF7B38F3AE57E8281122485EE4FC4B833A545523ED3EE4DAFDF9BEC8F7164"
)

var fontStruct = []struct {
	fontName string
	fontUrl  string
	fontSHA  string
}{
	{notoSansJPFontName, notoSansJPFontUrl, notoSansJPFontSHA256},
	{notoSansKRFontName, notoSansKRFontUrl, notoSansKRFontSHA256},
	{notoSansSCFontName, notoSansSCFontUrl, notoSansSCFontSHA256},
	{notoSansTCFontName, notoSansTCFontUrl, notoSansTCFontSHA256},
	{notoSansFontName, notoSansFontUrl, notoSansFontSHA256},
}

func InitFontsFiles() error {
	err := os.MkdirAll(FontsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create fonts directory: %w", err)
	}
	for _, font := range fontStruct {
		file, err := os.Open(FontsDir + font.fontName)
		if err != nil {
			log.Printf("Failed to open font file %s: %v", font.fontName, err)
			errDl := downloadFontFile(font.fontName, font.fontUrl, font.fontSHA)
			if errDl != nil {
				log.Printf("Failed to download font file %s: %v", font.fontName, errDl)
				return errDl
			}
			continue
		}
		hasher := sha256.New()
		_, copyErr := io.Copy(hasher, file)
		closeErr := file.Close()
		if copyErr != nil {
			return fmt.Errorf("failed to read font file %s: %w", font.fontName, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close font file %s: %w", font.fontName, closeErr)
		}
		hashBytes := hasher.Sum(nil)
		hashString := hex.EncodeToString(hashBytes)
		if !strings.EqualFold(hashString, font.fontSHA) {
			log.Printf("Font file %s has incorrect hash: expected %s, got %s. The file may have been TAMPERED with.", font.fontName, font.fontSHA, hashString)
			errDl := downloadFontFile(font.fontName, font.fontUrl, font.fontSHA)
			if errDl != nil {
				log.Printf("Failed to download font file %s: %v", font.fontName, errDl)
				return errDl
			}
		}
	}
	return nil
}

func downloadFontFile(fontName, fontUrl, fontSHA string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fontUrl, nil)
	if err != nil {
		return err
	}
	resp, err := HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download font file %s: status code %d", fontName, resp.StatusCode)
	}
	fontData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	if _, err := hasher.Write(fontData); err != nil {
		return err
	}
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	if !strings.EqualFold(hashString, fontSHA) {
		return fmt.Errorf("downloaded font file %s has incorrect hash: expected %s, got %s. The file may have been TAMPERED with.", fontName, fontSHA, hashString)
	}
	if err := os.WriteFile(FontsDir+fontName, fontData, 0644); err != nil {
		return err
	}
	return nil
}

func GetFontsFile(fontName string) ([]byte, error) {
	fontName = strings.TrimSuffix(fontName, ".ttf")
	fontName += ".zip"
	zipFilePath := FontsDir + fontName
	fontData, err := os.ReadFile(zipFilePath)
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
