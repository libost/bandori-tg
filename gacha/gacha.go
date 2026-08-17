package gacha

import (
	"encoding/json"
	"fmt"
	"net/http"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

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
