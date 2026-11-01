package characters

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"net/http"
	"time"

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

func getCharacterDetails(characterId string) (C.CharacterDetailed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	characterMap := utils.ReadCharacters()
	_, exists := characterMap[characterId]
	if !exists {
		return C.CharacterDetailed{}, fmt.Errorf("character not found")
	}

	url := "https://bestdori.com/api/characters/" + characterId + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return C.CharacterDetailed{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := utils.HttpClient.Do(req)
	if err != nil {
		return C.CharacterDetailed{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return C.CharacterDetailed{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var characterDetails C.CharacterDetailed
	err = json.UnmarshalRead(resp.Body, &characterDetails)
	if err != nil {
		return C.CharacterDetailed{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return characterDetails, nil
}

func convertAnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return ""
	}
}
