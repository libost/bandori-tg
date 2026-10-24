package gacha

import (
	"testing"
)

func TestGenerateGachaResultHonorsGuarantee(t *testing.T) {
	gachaMaps := []gachaMap{
		{Id: "2", WeightTotal: 1, Rarity: 2},
		{Id: "3", WeightTotal: 2, Rarity: 3},
		{Id: "4", WeightTotal: 3, Rarity: 4},
	}

	result, err := generateGachaResult(gachaMaps, 3)
	if err != nil {
		t.Fatalf("generateGachaResult returned an error: %v", err)
	}
	if result != "3" && result != "4" {
		t.Fatalf("generateGachaResult returned non-guaranteed card %q", result)
	}
}
