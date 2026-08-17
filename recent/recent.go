package recent

import (
	"strconv"

	"github.com/libost/bandori-tg/utils"
)

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
	recentData := utils.ReadRecent()
	keys := utils.ReadREventsKeys()
	var candidates []string
	for _, key := range keys {
		endAt := recentData.Events[key].EndAt
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
	latestValue, err := strconv.ParseInt(recentData.Events[latest].EndAt[region], 10, 64)
	if err != nil {
		return "", false
	}

	for _, key := range candidates[1:] {
		value, err := strconv.ParseInt(recentData.Events[key].EndAt[region], 10, 64)
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
