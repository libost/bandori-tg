package recent

import (
	"strconv"
	"time"

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

type GachaOngoingPerRegion struct {
	Region string
	Gacha  []string
}

// GetOngoingGachas returns a list of ongoing gachas, sorted by jp, en, tw, cn.
func GetOngoingGachas() ([]GachaOngoingPerRegion, error) {
	ongoing := make([]GachaOngoingPerRegion, 4)
	for region, index := range []int{0, 1, 2, 3} {
		gachaIDs := findLatestOngoingGachas(index)
		ongoing[region] = GachaOngoingPerRegion{
			Region: indexHelper(index),
			Gacha:  gachaIDs,
		}
	}
	return ongoing, nil
}

func findLatestOngoingGachas(region int) []string {
	recentData := utils.ReadRecent()
	keys := utils.ReadRGachasKeys()
	now := time.Now().Unix() * 1000
	var candidates []string
	for _, key := range keys {
		if _, ok := recentData.Gacha[key]; !ok {
			continue
		}
		endAt := recentData.Gacha[key].ClosedAt
		startAt := recentData.Gacha[key].PublishedAt
		if region >= len(endAt) || region >= len(startAt) {
			continue
		}
		value := endAt[region]
		valueStart := startAt[region]
		if value == "" || value == "null" || valueStart == "" || valueStart == "null" {
			continue
		}
		endTime, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		if endTime < now {
			continue
		}
		startTime, err := strconv.ParseInt(valueStart, 10, 64)
		if err != nil {
			continue
		}
		if startTime > now {
			continue
		}
		candidates = append(candidates, key)
	}

	return candidates
}

func indexHelper(index int) string {
	switch index {
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
