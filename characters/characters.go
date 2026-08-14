package characters

import (
	"fmt"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetCharacter(characterId string) (C.Character, error) {
	character, exists := utils.Characters[characterId]
	if !exists {
		return C.Character{}, fmt.Errorf("character not found")
	}
	return character, nil
}
