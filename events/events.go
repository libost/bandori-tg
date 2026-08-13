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

func GetDetailedEvents(eventID string) (C.EventDetailed, error) {
	resp, err := http.Get("https://bestdori.com/api/events/" + eventID + ".json")
	if err != nil {
		return C.EventDetailed{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return C.EventDetailed{}, err
	}
	var event C.EventDetailed
	err = json.Unmarshal([]byte(data), &event)
	if err != nil {
		return C.EventDetailed{}, err
	}
	return event, nil
}

func regionCodeFromEvent(event C.EventDetailed, preferCode string) string {
	index := 0
	switch preferCode {
	case "jp":
		index = 0
	case "en":
		index = 1
	case "tw":
		index = 2
	case "cn":
		index = 3
	case "kr":
		index = 4
	}
	if index < len(event.EndAt) && event.EndAt[index] != "" && event.EndAt[index] != "null" {
		return preferCode
	}
	for i, releasedAt := range event.EndAt {
		if releasedAt == "" || releasedAt == "null" {
			continue
		}
		switch i {
		case 0:
			return "jp"
		case 1:
			return "en"
		case 2:
			return "tw"
		case 3:
			return "cn"
		case 4:
			return "kr"
		default:
			return "jp"
		}
	}
	return "jp"
}
