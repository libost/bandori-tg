package constants

const (
	BandNameFile   = "./res/bandname.json"
	CardsFile      = "./res/cards.json"
	CharactersFile = "./res/characters.json"
	ConfigPath     = "./config.yaml"
	DatabaseFile   = "./res/bandori.db"
	EventsFile     = "./res/events.json"
	SkillsFile     = "./res/skills.json"
	RecentFile     = "./res/recent.json"
)

type Config struct {
	General struct {
		Token      string `yaml:"token"`
		AdminToken string `yaml:"admin_token"`
	} `yaml:"general"`
	Webhook struct {
		Enabled      bool   `yaml:"enabled"`
		NginxEnabled bool   `yaml:"nginx_enabled"`
		URL          string `yaml:"url"`
		Port         int    `yaml:"port"`
		Cert         struct {
			SelfSigned bool   `yaml:"self_signed"`
			CertPath   string `yaml:"cert_path"`
			KeyPath    string `yaml:"key_path"`
		} `yaml:"cert"`
		Secret string `yaml:"secret"`
	} `yaml:"webhook"`
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

type Recent struct {
	Songs      map[string]SongsData      `json:"songs"`
	Events     map[string]EventData      `json:"events"`
	Gacha      map[string]RGachaData     `json:"gacha"`
	LoginBonus map[string]LoginBonusData `json:"loginBonus"`
}

type SongsData struct {
	MusicTitle  []string `json:"musicTitle"`
	PublishedAt []string `json:"publishedAt"`
}

type EventData struct {
	EventName []string `json:"eventName"`
	StartAt   []string `json:"startAt"`
	EndAt     []string `json:"endAt"`
}

type RGachaData struct {
	GachaName   []string `json:"gachaName"`
	PublishedAt []string `json:"publishedAt"`
	ClosedAt    []string `json:"closedAt"`
}

type LoginBonusData struct {
	Caption     []string `json:"caption"`
	PublishedAt []string `json:"publishedAt"`
	ClosedAt    []string `json:"closedAt"`
}

type Events struct {
	EventType             string   `json:"eventType"`
	EventName             []string `json:"eventName"`
	AssetBundleName       string   `json:"assetBundleName"`
	BannerAssetBundleName string   `json:"bannerAssetBundleName"`
	StartAt               []string `json:"startAt"`
	EndAt                 []string `json:"endAt"`
	Attribute             []any    `json:"attribute"`
	Characters            []any    `json:"characters"`
	Members               []any    `json:"members"`
	LimitBreaks           []any    `json:"limitBreaks"`
	RewardCards           []any    `json:"rewardCards"`
}

type EventsData map[string]Events

type CardDetailed struct {
	CharacterID     int    `json:"characterId"`
	Rarity          int    `json:"rarity"`
	Attribute       string `json:"attribute"`
	LevelLimit      int    `json:"levelLimit"`
	ResourceSetName string `json:"resourceSetName"`
	SdResourceName  string `json:"sdResourceName"`
	//Episodes        map[string]CardStatValue `json:"episodes"` //???
	CostumeID  int      `json:"costumeId"`
	GachaText  []string `json:"gachaText"`
	Prefix     []string `json:"prefix"`
	ReleasedAt []string `json:"releasedAt"`
	SkillName  []string `json:"skillName"`
	SkillID    int      `json:"skillId"`
	Source     []any    `json:"source"`
	Type       string   `json:"type"`
	//Stats           CardStatValue            `json:"stats"`
}

type EventDetailed struct {
	EventType                       string            `json:"eventType"`
	EventName                       []string          `json:"eventName"`
	AssetBundleName                 string            `json:"assetBundleName"`
	BannerAssetBundleName           string            `json:"bannerAssetBundleName"`
	StartAt                         []string          `json:"startAt"`
	EndAt                           []string          `json:"endAt"`
	EnableFlag                      []*bool           `json:"enableFlag"`
	PublicStartAt                   []string          `json:"publicStartAt"`
	PublicEndAt                     []string          `json:"publicEndAt"`
	DistributionStartAt             []string          `json:"distributionStartAt"`
	DistributionEndAt               []string          `json:"distributionEndAt"`
	BgmAssetBundleName              string            `json:"bgmAssetBundleName"`
	BgmFileName                     string            `json:"bgmFileName"`
	AggregateEndAt                  []string          `json:"aggregateEndAt"`
	ExchangeEndAt                   []string          `json:"exchangeEndAt"`
	PointRewards                    []any             `json:"pointRewards"`
	RankingRewards                  []any             `json:"rankingRewards"`
	Attributes                      []EventAttribute  `json:"attributes"`
	Characters                      []EventCharacters `json:"characters"`
	EventAttributeAndCharacterBonus struct {
		EventID          int `json:"eventId"`
		PointPercent     int `json:"pointPercent"`
		ParameterPercent int `json:"parameterPercent"`
	} `json:"eventAttributeAndCharacterBonus"`
	Members     []EventMembers     `json:"members"`
	LimitBreaks []EventLimitBreaks `json:"limitBreaks"`
	Stories     []any              `json:"stories"`
	RewardCards []int              `json:"rewardCards"`
}

type EventAttribute struct {
	EventID   int    `json:"eventId"`
	Attribute string `json:"attribute"`
	Percent   int    `json:"percent"`
}

type EventCharacters struct {
	EventID     int `json:"eventId"`
	CharacterID int `json:"characterId"`
	Percent     int `json:"percent"`
	Seq         int `json:"seq"`
}

type EventMembers struct {
	EventID     int `json:"eventId"`
	SituationID int `json:"situationId"`
	Percent     int `json:"percent"`
	Seq         int `json:"seq"`
}

type EventLimitBreaks struct {
	Rarity  int     `json:"rarity"`
	Rank    int     `json:"rank"`
	Percent float32 `json:"percent"`
}

type Gacha struct {
	ResourceName          string   `json:"resourceName"`
	BannerAssetBundleName string   `json:"bannerAssetBundleName"`
	GachaName             []string `json:"gachaName"`
	PublishedAt           []string `json:"publishedAt"`
	ClosedAt              []string `json:"closedAt"`
	Type                  string   `json:"type"`
	NewCards              []int    `json:"newCards"`
}

type GachaData map[string]Gacha

type GachaDetailed struct {
	Details               []any            `json:"details"`
	Rates                 []any            `json:"rates"`
	PaymentMethods        []PaymentMethods `json:"paymentMethods"`
	ResourceName          string           `json:"resourceName"`
	BannerAssetBundleName string           `json:"bannerAssetBundleName"`
	GachaName             []string         `json:"gachaName"`
	PublishedAt           []string         `json:"publishedAt"`
	ClosedAt              []string         `json:"closedAt"`
	Description           []string         `json:"description"`
	Annotation            []string         `json:"annotation"`
	GachaPeriod           []string         `json:"gachaPeriod"`
	GachaType             string           `json:"gachaType"`
	Type                  string           `json:"type"`
	NewCards              []int            `json:"newCards"`
	Information           struct {
		Description   []string `json:"description"`
		Term          []string `json:"term"`
		NewMemberInfo []string `json:"newMemberInfo"`
		Notice        []string `json:"notice"`
	} `json:"information"`
}

type PaymentMethods struct {
	GachaID          string `json:"gachaId"`
	PaymentMethod    string `json:"paymentMethod"`
	Quantity         int    `json:"quantity"`
	PaymentMethodID  int    `json:"paymentMethodId"`
	Count            int    `json:"count"`
	Behavior         string `json:"behavior"`
	Pickup           bool   `json:"pickup"`
	CostItemQuantity int    `json:"costItemQuantity"`
	DiscountType     string `json:"discountType"`
}

// Access emoji by using CharaEmoji[characterID-1], for example, CharaEmoji[0] will return the emoji for character ID 1.
var CharaEmoji = [...]int64{
	6046130500299398386,
	6046387158955072399,
	6046167522917490460,
	6046585002328595828,
	6046391398087794995,
	6046433647681085078,
	6046499274781369282,
	6048742759538368517,
	6046090015937665512,
	6046150751070199791,
	6046307225318728001,
	6046552386346951203,
	6048708541533919853,
	6046364180880040403,
	6048618806782205656,
	6046638311462674031,
	6046479015420632485,
	6046158061104536628,
	6046380089438904123,
	6046340653049192832,
	6048538409289392417,
	6046637108871831280,
	6046188512422665882,
	6046526560708599605,
	6048779756386655755,
	6048523166450458769,
	6046397548480962102,
	6046232248074641192,
	6046236272458997711,
	6046423262450164145,
	6046153035992801227,
	6046161514258242737,
	6048380341607996514,
	6046462896408371140,
	6046380372906746121,
	6046297857995054199,
	6046085957193572423,
	6048654751363505392,
	6046456028755664146,
	6048780769998942834,
}

// Access emoji by using BandEmoji[bandID-1], for example, BandEmoji[0] will return the emoji for band ID 1.
// For unknown reasons, the IDs of RAS, Morfonica and MyGO are 18, 21 and 45 respectively, so the missing IDs will be filled with 0.
var BandEmoji = [...]int64{
	6048882135522089606,
	6046584078910627512,
	6046608500094672312,
	6046308900355973792,
	6046508027924717838,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	6046376782314085971, // 18
	0,
	0,
	6046439415822166401, // 21
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	0,
	6046202260612981011, // 45
}

// Access emoji by using AttributeEmoji[attributeIndex], for example, AttributeEmoji[0] will return the emoji for attribute "powerful".
var AttributeEmoji = [...]int64{
	6046234846529854324, // powerful
	6048435136800761010, // cool
	6046477052620578270, // happy
	6046633359365382486, // pure
}

// Access emoji by using RarityEmoji[rarityIndex], for example, RarityEmoji[0] will return the emoji for rarity "normal".
var RarityEmoji = [...]int64{
	6048893384041438933, // Normal
	6048597808687095847, // After Training
}

var AcceptedRegions = []string{"jp", "en", "tw", "kr", "cn"}

// These cards are known to have issues with their data, so they are excluded from the bot's responses.
var BadCards = []string{"2309"}

var TelegramAcceptPorts = []int{80, 88, 443, 8443}
