package recent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var Recent C.Recent

var REventsKeys []string

func InitLists() error {
	if err := os.MkdirAll("./res", 0755); err != nil {
		return err
	}
	timedRefresh()
	return nil
}

func timedRefresh() {
	go func() {
		refreshLists()
		UnmarshalList()
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
	fmt.Println("Recent list refreshed successfully.")
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
	for region, index := range []int{0, 1, 2, 3} {
		latestID, ok := findLatestOngoingEvent(index)
		if ok {
			ongoing[region] = latestID
		}
	}
	return ongoing, nil
}

func findLatestOngoingEvent(region int) (string, bool) {
	var candidates []string
	for _, key := range REventsKeys {
		endAt := Recent.Events[key].EndAt
		if region >= len(endAt) {
			continue
		}
		value := endAt[region]
		if value == "" || value == "null" {
			continue
		}
		candidates = append(candidates, key)
	}
	if len(candidates) == 0 {
		return "", false
	}

	latest := candidates[0]
	latestValue, err := strconv.ParseInt(Recent.Events[latest].EndAt[region], 10, 64)
	if err != nil {
		return "", false
	}

	for _, key := range candidates[1:] {
		value, err := strconv.ParseInt(Recent.Events[key].EndAt[region], 10, 64)
		if err != nil {
			continue
		}
		if value > latestValue {
			latest = key
			latestValue = value
		}
	}

	return latest, true
}
