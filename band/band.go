package band

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	C "github.com/libost/bandori-tg/constants"
)

var Band map[string]C.Band

func InitLists() error {
	if err := os.MkdirAll(filepath.Dir(C.BandNameFile), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(C.BandNameFile); os.IsNotExist(err) {
		err := refreshLists()
		if err != nil {
			return err
		}
	}
	return UnmarshalList()
}

func refreshLists() error {
	resp, err := http.Get("https://bestdori.com/api/bands/main.1.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.BandNameFile)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// Use bandname.Band["bandId"] to access band name, for example, bandname.Band["1"].BandName[`int`] will return the band name for the band with ID 1.
func UnmarshalList() error {
	file, err := os.Open(C.BandNameFile)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Band)
	if err != nil {
		return err
	}
	return nil
}

func GetBand(bandId string) (C.Band, error) {
	band, exists := Band[bandId]
	if !exists {
		return C.Band{}, fmt.Errorf("band with ID %s not found", bandId)
	}
	return band, nil
}
