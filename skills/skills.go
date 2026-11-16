package skills

import (
	"fmt"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetSkill(skillID string) (C.Skills, error) {
	skillMap := utils.ReadSkills()
	skill, exists := skillMap[skillID]
	if !exists {
		return C.Skills{}, fmt.Errorf("skill with ID %s not found", skillID)
	}
	return skill, nil
}
