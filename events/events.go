package events

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var Events C.EventsData

func InitLists() error {
	if err := os.MkdirAll("./res", 0755); err != nil {
		return err
	}
	if err := refreshLists(); err != nil {
		return err
	}
	if err := UnmarshalList(); err != nil {
		return err
	}
	timedRefresh()
	return nil
}

func timedRefresh() {
	go func() {
		ticker := time.NewTicker(3 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refreshLists()
			UnmarshalList()
		}
	}()
}

func refreshLists() error {
	resp, err := http.Get("https://bestdori.com/api/events/all.5.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.EventsFile)
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

func UnmarshalList() error {
	file, err := os.Open(C.EventsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &Events)
	if err != nil {
		return err
	}
	return nil
}
