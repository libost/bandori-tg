package gacha

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"math/rand/v2"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/libost/bandori-tg/cards"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

type gachaMap struct {
	Id          string
	Weight      int
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

func generateGachaResultByRates(gachaMaps []gachaMap, rarityWeights []int, guarantee int) (string, error) {
	if len(gachaMaps) == 0 {
		return "", fmt.Errorf("no gacha cards available")
	}

	rarityTotal := 0
	eligibleRarities := make([]int, 0, len(rarityWeights))
	for rarity, weight := range rarityWeights {
		rarityIndex := rarity + 1
		if rarityIndex < guarantee || weight <= 0 {
			continue
		}
		hasCards := false
		for _, gacha := range gachaMaps {
			if gacha.Rarity == rarityIndex && gacha.Weight > 0 {
				hasCards = true
				break
			}
		}
		if hasCards {
			rarityTotal += weight
			eligibleRarities = append(eligibleRarities, rarityIndex)
		}
	}
	if len(eligibleRarities) == 0 {
		return "", fmt.Errorf("no gacha rarities meet rarity guarantee: %d", guarantee)
	}

	randomNum, err := randomNumGen(1, rarityTotal)
	if err != nil {
		return "", err
	}
	selectedRarity := 0
	for _, rarity := range eligibleRarities {
		randomNum -= rarityWeights[rarity-1]
		if randomNum <= 0 {
			selectedRarity = rarity
			break
		}
	}
	if selectedRarity == 0 {
		return "", fmt.Errorf("failed to generate gacha rarity result")
	}

	cardWeightTotal := 0
	for _, gacha := range gachaMaps {
		if gacha.Rarity == selectedRarity && gacha.Weight > 0 {
			cardWeightTotal += gacha.Weight
		}
	}
	randomNum, err = randomNumGen(1, cardWeightTotal)
	if err != nil {
		return "", err
	}
	for _, gacha := range gachaMaps {
		if gacha.Rarity != selectedRarity || gacha.Weight <= 0 {
			continue
		}
		randomNum -= gacha.Weight
		if randomNum <= 0 {
			return gacha.Id, nil
		}
	}
	return "", fmt.Errorf("failed to generate gacha card result")
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
	if index >= len(gachaDetailed.Details) || index >= len(gachaDetailed.Rates) {
		return nil, fmt.Errorf("no gacha data found for region code: %s", regionCode)
	}
	details := gachaDetailed.Details[index]
	if len(details) == 0 {
		return nil, fmt.Errorf("no gacha details found for region code: %s", regionCode)
	}
	var gachaMaps []gachaMap
	for id, gacha := range details {
		gachaMaps = append(gachaMaps, gachaMap{
			Id:     id,
			Weight: gacha.Weight,
			Rarity: gacha.RarityIndex,
		})
	}
	var rarityRates [5]float32

	for rarity, rate := range gachaDetailed.Rates[index] {
		var rarityIndex int
		if _, err := fmt.Sscanf(rarity, "%d", &rarityIndex); err != nil {
			return nil, fmt.Errorf("invalid rarity %q: %w", rarity, err)
		}

		if rarityIndex < 1 || rarityIndex > len(rarityRates) {
			return nil, fmt.Errorf("rarity out of range: %d", rarityIndex)
		}

		rarityRates[rarityIndex-1] = rate.Rate
	}
	rarityWeightsInt := make([]int, len(rarityRates))
	for i, rate := range rarityRates {
		rarityWeightsInt[i] = int(math.Round(float64(rate) * 100))
	}
	guaranteedRarity := 3
	switch paymentMethods.Behavior {
	case "fixed_5_star_once":
		guaranteedRarity = 5
	}
	hasFulfilledGuarantee := false
	for i := range pull {
		if i == pull-1 && pull != 1 && !hasFulfilledGuarantee {

			result, err := generateGachaResultByRates(gachaMaps, rarityWeightsInt, guaranteedRarity)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		} else {
			result, err := generateGachaResultByRates(gachaMaps, rarityWeightsInt, 1)
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
