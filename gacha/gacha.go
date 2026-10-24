package gacha

import (
	"encoding/json"
	"fmt"
	"image"
	"math/rand/v2"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/libost/bandori-tg/cards"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

type gachaMap struct {
	Id          string
	WeightTotal int
	Rarity      int
}

func GetGacha(gachaId string) (C.Gacha, error) {
	gachaMap := utils.ReadGacha()
	gacha, exists := gachaMap[gachaId]
	if !exists {
		return C.Gacha{}, fmt.Errorf("gacha with ID %s not found", gachaId)
	}
	return gacha, nil
}

func GetGachaDetails(gachaId string) (C.GachaDetailed, error) {
	gachaMap := utils.ReadGacha()
	_, exists := gachaMap[gachaId]
	if !exists {
		return C.GachaDetailed{}, fmt.Errorf("gacha with ID %s not found", gachaId)
	}
	resp, err := http.Get("https://bestdori.com/api/gacha/" + gachaId + ".json")
	if err != nil {
		return C.GachaDetailed{}, err
	}
	defer resp.Body.Close()
	var gachaDetailed C.GachaDetailed
	if err := json.NewDecoder(resp.Body).Decode(&gachaDetailed); err != nil {
		return C.GachaDetailed{}, err
	}
	return gachaDetailed, nil
}

func generateGachaResult(gachaMaps []gachaMap, guarantee int) (string, error) {
	if len(gachaMaps) == 0 {
		return "", fmt.Errorf("no gacha cards available")
	}

	eligible := make([]gachaMap, 0, len(gachaMaps))
	weightTotal := 0
	previousWeightTotal := 0
	for _, gacha := range gachaMaps {
		weight := gacha.WeightTotal - previousWeightTotal
		previousWeightTotal = gacha.WeightTotal
		if gacha.Rarity >= guarantee {
			if weight <= 0 {
				continue
			}
			weightTotal += weight
			eligible = append(eligible, gachaMap{
				Id:          gacha.Id,
				WeightTotal: weightTotal,
				Rarity:      gacha.Rarity,
			})
		}
	}
	if len(eligible) == 0 {
		return "", fmt.Errorf("no gacha cards meet rarity guarantee: %d", guarantee)
	}

	randomNum, err := randomNumGen(1, weightTotal)
	if err != nil {
		return "", err
	}
	for _, gacha := range eligible {
		if randomNum <= gacha.WeightTotal {
			return gacha.Id, nil
		}
	}
	return "", fmt.Errorf("failed to generate gacha result")
}

func randomNumGen(min int, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	return rand.N(max-min+1) + min, nil
}

func gachaHelper(gachaDetailed C.GachaDetailed, regionCode string, paymentMethods C.PaymentMethods) ([]string, error) {
	pull := paymentMethods.Count
	if pull <= 0 {
		return nil, fmt.Errorf("pull count must be greater than zero")
	}

	var results []string
	index := indexHelper(regionCode)
	details := gachaDetailed.Details[index]
	if len(details) == 0 {
		return nil, fmt.Errorf("no gacha details found for region code: %s", regionCode)
	}
	var gachaMaps []gachaMap
	weightTotal := 0
	for id, gacha := range details {
		weightTotal += gacha.Weight
		gachaMaps = append(gachaMaps, gachaMap{
			Id:          id,
			WeightTotal: weightTotal,
			Rarity:      gacha.RarityIndex,
		})
	}
	guaranteedRarity := 3
	switch paymentMethods.Behavior {
	case "fixed_5_star_once":
		guaranteedRarity = 5
	}
	hasFulfilledGuarantee := false
	for i := range pull {
		if i == pull-1 && pull != 1 && !hasFulfilledGuarantee {
			result, err := generateGachaResult(gachaMaps, guaranteedRarity)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		} else {
			result, err := generateGachaResult(gachaMaps, 0)
			if err != nil {
				return nil, err
			}
			gacha := details[result]
			if gacha.RarityIndex >= guaranteedRarity {
				hasFulfilledGuarantee = true
			}
			results = append(results, result)
		}
	}
	return results, nil
}

func generateGachaImage(results []string) (image.Image, error) {
	dc := gg.NewContext(len(results)*180+20*len(results)+20, 300)
	dc.SetHexColor("#FFFFFF")
	dc.Clear()
	for i, result := range results {
		cardImage, err := cards.ThumbGenerate(result, 0)
		if err != nil {
			return nil, err
		}
		dc.DrawImageAnchored(cardImage, 200*i+20, 25, 0, 0)
	}
	return dc.Image(), nil
}
