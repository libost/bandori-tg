package band

import (
	"fmt"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetBand(bandId string) (C.Band, error) {
	bandMap := utils.ReadBand()
	band, exists := bandMap[bandId]
	if !exists {
		return C.Band{}, fmt.Errorf("band with ID %s not found", bandId)
	}
	return band, nil
}
