package cards

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var Cards C.CardData

func InitLists() error {
	if err := os.MkdirAll(filepath.Dir(C.CardsFile), 0755); err != nil {
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
	resp, err := http.Get("https://bestdori.com/api/cards/all.5.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.CardsFile)
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

type cardStat C.CardStat

func (s *cardStat) UnmarshalJSON(data []byte) error {
	// 第一步：先将 stat 解析为一个原始的 Map，Value 保持为 json.RawMessage (延迟解析)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Levels = make(map[string]C.CardStatValue)

	// 第二步：遍历这个 Map，区分处理 "episodes" 和 其它等级 Key
	for key, rawVal := range raw {
		if key == "episodes" {
			// 如果 key 是 "episodes"，按 []StatValue 数组解析
			if err := json.Unmarshal(rawVal, &s.Episodes); err != nil {
				return err
			}
		} else {
			// 剩下的 key ("1", "20" 等)，按单条 StatValue 解析
			var val C.CardStatValue
			if err := json.Unmarshal(rawVal, &val); err != nil {
				return err
			}
			s.Levels[key] = val
		}
	}
	return nil
}

// Use cards.Cards["cardId"] to access card data, for example, cards.Cards["1"] will return the card data for the card with ID 1.
// For `stats`, you can access the stats for a specific level using `cards.Cards["1"].Stat.Levels["1"]` for level 1 stats, or `cards.Cards["1"].Stat.Episodes[int]` for episode stats.
func UnmarshalList() error {
	file, err := os.Open(C.CardsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &Cards)
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}
	return nil
}

func cardOrdinaryType(types string) bool {
	return types == "permanent" || types == "limited" || types == "dreamfes" || types == "event"
}

func regionCodeFromCard(card C.Card) string {
	for i, releasedAt := range card.ReleasedAt {
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

// saveCardsToFile saves the card image to a file and returns the file path.
// regionCode: "en" for English, "jp" for Japanese, etc.
// resourceSetName: The resource set name of the card.
// cardStat: The card stat, e.g., "normal", "after_training".
func retrieveCardsPath(regionCode string, resourceSetName string, cardStat string) (string, error) {
	// Implementation for retrieving card image paths
	url := "https://bestdori.com/assets/" + regionCode + "/characters/resourceset/" + resourceSetName + "_rip/card_" + cardStat + ".png"
	return url, nil
}

func GetCard(cardId string) (string, string, error) {
	card, exists := Cards[cardId]
	if !exists {
		return "", "", nil
	}
	regionCode := regionCodeFromCard(card)
	if card.Rarity < 3 || !cardOrdinaryType(card.Type) {
		if card.Type == "campaign" {
			if card.Rarity < 3 {
				normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
				if err != nil {
					return "", "", err
				}
				return normalPath, "", nil
			}
			trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
			if err != nil {
				return "", "", err
			}
			return "", trainingPath, nil
		}
		if card.Rarity < 3 {

			normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
			if err != nil {
				return "", "", err
			}
			return normalPath, "", nil
		}
		if card.Type == "birthday" || card.Type == "others" || card.Type == "kirafes" {
			trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
			if err != nil {
				return "", "", err
			}
			return "", trainingPath, nil
		}

	}
	normalPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "normal")
	if err != nil {
		return "", "", err
	}
	trainingPath, err := retrieveCardsPath(regionCode, card.ResourceSetName, "after_training")
	if err != nil {
		return "", "", err
	}
	return normalPath, trainingPath, nil
}

func GetDetailedCard(cardId string) (C.CardDetailed, error) {
	resp, err := http.Get("https://bestdori.com/api/cards/" + cardId + ".json")
	if err != nil {
		return C.CardDetailed{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return C.CardDetailed{}, err
	}
	var cardDetailed C.CardDetailed
	err = json.Unmarshal(body, &cardDetailed)
	if err != nil {
		return C.CardDetailed{}, err
	}
	return cardDetailed, nil
}

/*
func GetGachaVoice(resourceSetName string, cardType string, regionCode string) (string, error) {
	switch cardType {
	case "birthday":
		cardType = "birthday"
	case "permanent":
		cardType = "operation"
	case "limited", "dreamfes":
		cardType = "limited"
	}
	url := "https://bestdori.com/assets/" + regionCode + "/gacha/voice/" + cardType + "/spin_rip/" + resourceSetName + ".mp3"
	return url, nil
}
*/
