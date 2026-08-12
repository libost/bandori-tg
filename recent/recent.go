package recent

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var Recent C.Recent

var REventsKeys []string

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
	resp, err := http.Get("https://bestdori.com/api/news/dynamic/recent.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.RecentFile)
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
	file, err := os.Open(C.RecentFile)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &Recent)
	if err != nil {
		return err
	}
	REventsKeys = make([]string, 0, len(Recent.Events))
	for k := range Recent.Events {
		REventsKeys = append(REventsKeys, k)
	}
	sort.Strings(REventsKeys)
	return nil
}

// GetOngoingEventsID returns a list of ongoing eventsIDs, sorted by jp, en, tw, cn.
func GetOngoingEventsID() ([]string, error) {
	ongoing := make([]string, 4)
	ongoing[0] = REventsKeys[len(REventsKeys)-1]
	ongoing[1] = REventsKeys[len(REventsKeys)-7]
	ongoing[2] = REventsKeys[len(REventsKeys)-3]
	ongoing[3] = REventsKeys[len(REventsKeys)-5]
	return ongoing, nil
}
