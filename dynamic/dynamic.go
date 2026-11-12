package dynamic

import "strconv"

func shouldCollectSong(publishedAt []string, idx int) bool {
	if idx < 0 || idx >= len(publishedAt) || publishedAt[idx] == "" {
		return false
	}

	currentPubTime, err := strconv.ParseInt(publishedAt[idx], 10, 64)
	if err != nil {
		return false
	}

	validValueCount := 0
	for _, value := range publishedAt {
		if value == "" {
			continue
		}
		pubTime, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}
		validValueCount++
		if pubTime > currentPubTime {
			return false
		}
	}

	return validValueCount > 0
}
