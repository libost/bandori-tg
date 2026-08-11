package characters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	C "github.com/libost/bandori-tg/constants"
)

var Characters map[string]C.Character

func InitLists() error {
	if err := os.MkdirAll(filepath.Dir(C.CharactersFile), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(C.CharactersFile); os.IsNotExist(err) {
		err := refreshLists()
		if err != nil {
			return err
		}
	}
	return UnmarshalList()
}

func refreshLists() error {
	resp, err := http.Get("https://bestdori.com/api/characters/main.3.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.CharactersFile)
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

// Use characters.Characters["characterId"] to access character data, for example, characters.Characters["1"] will return the character data for the character with ID 1.
func UnmarshalList() error {
	file, err := os.Open(C.CharactersFile)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Characters)
	if err != nil {
		return err
	}
	return nil
}

func GetCharacter(characterId string) (C.Character, error) {
	character, exists := Characters[characterId]
	if !exists {
		return C.Character{}, fmt.Errorf("character not found")
	}
	return character, nil
}
