package cards

import (
	"encoding/json"
	"io"
	"net/http"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func cardOrdinaryType(types string) bool {
	return types == "permanent" || types == "limited" || types == "dreamfes" || types == "event"
}

func regionCodeFromCard(card C.Card, preferCode string) string {
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
	if index < len(card.ReleasedAt) && card.ReleasedAt[index] != "" && card.ReleasedAt[index] != "null" {
		return preferCode
	}
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

// retrieveCardsPath retrieves the path for a card image based on its region, resource set name, and stat, then returns its url.
// regionCode: "en" for English, "jp" for Japanese, etc.
// resourceSetName: The resource set name of the card.
// cardStat: The card stat, e.g., "normal", "after_training".
func retrieveCardsPath(regionCode string, resourceSetName string, cardStat string) (string, error) {
	// Implementation for retrieving card image paths
	url := "https://bestdori.com/assets/" + regionCode + "/characters/resourceset/" + resourceSetName + "_rip/card_" + cardStat + ".png"
	return url, nil
}

func GetCard(cardId string) (string, string, error) {
	cardMap := utils.ReadCards()
	card, exists := cardMap[cardId]
	if !exists {
		return "", "", nil
	}
	regionCode := regionCodeFromCard(card, "jp")
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
