package characters

import (
	"fmt"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetCharacter(characterId string) (C.Character, error) {
	characterMap := utils.ReadCharacters()
	character, exists := characterMap[characterId]
	if !exists {
		return C.Character{}, fmt.Errorf("character not found")
	}
	return character, nil
}
