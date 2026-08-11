package constants

const (
	BandNameFile   = "./res/bandname.json"
	CardsFile      = "./res/cards.json"
	CharactersFile = "./res/characters.json"
	ConfigPath     = "./config.yaml"
	DatabaseFile   = "./res/bandori.db"
	SkillsFile     = "./res/skills.json"
)

type Config struct {
	General struct {
		Token string `yaml:"token"`
	} `yaml:"general"`
}

type Character struct {
	CharacterType string    `json:"characterType"`
	CharacterName []string  `json:"characterName"`
	FirstName     []string  `json:"firstName"`
	LastName      []string  `json:"lastName"`
	Nickname      []*string `json:"nickname"`
	BandID        int       `json:"bandId"`
	ColorCode     string    `json:"colorCode"`
}

type CharacterData map[string]Character

type Card struct {
	CharacterID     int      `json:"characterId"`
	Rarity          int      `json:"rarity"`
	Attribute       string   `json:"attribute"`
	LevelLimit      int      `json:"levelLimit"`
	ResourceSetName string   `json:"resourceSetName"`
	Prefix          []string `json:"prefix"`
	ReleasedAt      []string `json:"releasedAt"`
	SkillID         int      `json:"skillId"`
	Type            string   `json:"type"`
	Stat            CardStat `json:"stat"`
}

type CardStatValue struct {
	Performance int `json:"performance"`
	Technique   int `json:"technique"`
	Visual      int `json:"visual"`
}

type CardStat struct {
	Levels   map[string]CardStatValue `json:"-"`
	Episodes []CardStatValue          `json:"episodes"`
}

type CardData map[string]Card

type Skill struct {
	SimpleDescription []string  `json:"simpleDescription"`
	Description       []string  `json:"description"`
	Duration          []float32 `json:"duration"`
	ActivationEffect  struct {
		ActivateEffectTypes struct {
			Score SkillActivationEffectTypesValue `json:"score"`
			Judge SkillActivationEffectTypesValue `json:"judge"`
		}
	} `json:"activationEffect"`
	OnceEffect struct {
		OnceEffectType      string `json:"onceEffectType"`
		OnceEffectValueType string `json:"onceEffectValueType"`
		OnceEffectValue     []int  `json:"onceEffectValue"`
	} `json:"onceEffect"`
}

type SkillActivationEffectTypesValue struct {
	ActivateEffectValue     []int  `json:"activateEffectValue"`
	ActivateEffectValueType string `json:"activateEffectValueType"`
	ActivateCondition       string `json:"activateCondition"`
}

type SkillData map[string]Skill

type Band struct {
	BandName []string `json:"bandName"`
}

type BandData map[string]Band
