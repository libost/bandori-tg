package comic

import (
	"fmt"
	"strings"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetComicData(comicID string) (C.Comic, error) {
	comicData := utils.ReadComic()
	if comic, exists := comicData[comicID]; exists {
		return comic, nil
	}
	return C.Comic{}, fmt.Errorf("comic not found: %s", comicID)
}

func GetAllComicIDsForSpecifiedRegion(region string) []string {
	comicData := utils.ReadComic()
	idx := indexHelper(region)
	comicIDs := make([]string, 0, len(comicData))
	for comicID, comic := range comicData {
		publicStartAt := convertAnyToInt(comic.PublicStartAt[idx])
		if publicStartAt != 0 && comic.Title[idx] != "" {
			comicIDs = append(comicIDs, comicID)
		}
	}
	return comicIDs
}

func getComicUrl(comic C.Comic, region string) (string, error) {
	idx := indexHelper(region)
	if comic.PublicStartAt[idx] == 0 || comic.Title[idx] == "" {
		return "", fmt.Errorf("comic not available in the specified region: %s", region)
	}
	parts := strings.Split(comic.AssetBundleName, "_")
	if parts[0] != "comic" {
		return "", fmt.Errorf("invalid asset bundle name: %s", comic.AssetBundleName)
	}
	if len(parts) > 3 || len(parts) < 2 {
		return "", fmt.Errorf("invalid asset bundle name format: %s", comic.AssetBundleName)
	}
	if parts[1] == "fourframe" {
		return fmt.Sprintf("https://bestdori.com/assets/%s/comic/comic_fourframe/%s_rip/%s.png", region, comic.AssetBundleName, comic.AssetBundleName), nil
	}
	return fmt.Sprintf("https://bestdori.com/assets/%s//comic/comic_singleframe/%s_rip/%s.png", region, comic.AssetBundleName, comic.AssetBundleName), nil
}

func indexHelper(region string) int {
	switch region {
	case "jp":
		return 0
	case "en":
		return 1
	case "tw":
		return 2
	case "cn":
		return 3
	case "kr":
		return 4
	default:
		return 0
	}
}

func reverseIndexHelper(idx int) string {
	switch idx {
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

func convertAnyToInt(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case string:
		var intValue int
		_, err := fmt.Sscanf(v, "%d", &intValue)
		if err != nil {
			return 0
		}
		return intValue
	default:
		return 0
	}
}
