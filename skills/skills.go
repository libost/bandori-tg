package skills

import (
	"fmt"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/utils"
)

func GetSkill(skillID string) (C.Skill, error) {
	skill, exists := utils.Skills[skillID]
	if !exists {
		return C.Skill{}, fmt.Errorf("Skill with ID %s not found", skillID)
	}
	return skill, nil
}
