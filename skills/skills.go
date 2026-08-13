package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	C "github.com/libost/bandori-tg/constants"
)

var Skills C.SkillData

func InitLists() error {
	if err := os.MkdirAll(filepath.Dir(C.SkillsFile), 0755); err != nil {
		return err
	}
	if err := refreshLists(); err != nil {
		return err
	}
	if err := UnmarshalList(); err != nil {
		return err
	}
	timedRefresh()
	return nil
}

func timedRefresh() {
	go func() {
		ticker := time.NewTicker(3 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			refreshLists()
			UnmarshalList()
		}
	}()
}

func refreshLists() error {
	resp, err := http.Get("https://bestdori.com/api/skills/all.10.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Write the response body to a file
	out, err := os.Create(C.SkillsFile)
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

// Use skills.Skills["skillID"] to access skill data
func UnmarshalList() error {
	file, err := os.Open(C.SkillsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &Skills)
	if err != nil {
		log.Fatalf("解析失败: %v", err)
	}
	return nil
}

func GetSkill(skillID string) (C.Skill, error) {
	skill, exists := Skills[skillID]
	if !exists {
		return C.Skill{}, fmt.Errorf("Skill with ID %s not found", skillID)
	}
	return skill, nil
}
