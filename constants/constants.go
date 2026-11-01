package constants

import (
	"fmt"

	"github.com/sahilm/fuzzy"
)

const (
	BaseDir      = "./res/"
	ConfigPath   = "./config.yaml"
	DatabaseFile = "./res/bandori.db"
)

var (
	ErrCannotPredict = fmt.Errorf("not enough data to make predictions")
	ErrNoCutoffData  = fmt.Errorf("no cutoff data available for this event")
	ErrNoSuchEvent   = fmt.Errorf("no such event")
)

var (
	ErrNoBotToken    = fmt.Errorf("no bot token was provided in the config file")
	ErrNoAdminToken  = fmt.Errorf("no admin token was provided in the config file")
	ErrNoPicDepot    = fmt.Errorf("no pic depot was provided in the config file")
	ErrNoCertPath    = fmt.Errorf("self-signed certificate is enabled, but no certificate or key path was provided in the config file")
	ErrNoWebhookURL  = fmt.Errorf("webhook is enabled, but no webhook URL was provided in the config file")
	ErrNoWebhookPort = fmt.Errorf("webhook is enabled, but no webhook port was provided in the config file")
)

const (
	Card1PowerfulFrame = "https://bestdori.com/res/image/card-1-powerful.png"
	Card1CoolFrame     = "https://bestdori.com/res/image/card-1-cool.png"
	Card1HappyFrame    = "https://bestdori.com/res/image/card-1-happy.png"
	Card1PureFrame     = "https://bestdori.com/res/image/card-1-pure.png"
	Card2Frame         = "https://bestdori.com/res/image/card-2.png"
	Card3Frame         = "https://bestdori.com/res/image/card-3.png"
	Card4Frame         = "https://bestdori.com/res/image/card-4.png"
	Card5Frame         = "https://bestdori.com/res/image/card-5.png"
	PowerfulFrame      = "https://bestdori.com/res/icon/powerful.svg"
	CoolFrame          = "https://bestdori.com/res/icon/cool.svg"
	HappyFrame         = "https://bestdori.com/res/icon/happy.svg"
	PureFrame          = "https://bestdori.com/res/icon/pure.svg"
	StarIcon           = "https://bestdori.com/res/icon/star.png"
)

var Card1Frames = [...]string{
	Card1PowerfulFrame,
	Card1CoolFrame,
	Card1HappyFrame,
	Card1PureFrame,
}

var CardFrames = [...]string{
	Card2Frame,
	Card3Frame,
	Card4Frame,
	Card5Frame,
}

var AttributeFrames = [...]string{
	PowerfulFrame,
	CoolFrame,
	HappyFrame,
	PureFrame,
}

var BandFrames = [...]string{
	"https://bestdori.com/res/icon/band_1.svg",
	"https://bestdori.com/res/icon/band_2.svg",
	"https://bestdori.com/res/icon/band_3.svg",
	"https://bestdori.com/res/icon/band_4.svg",
	"https://bestdori.com/res/icon/band_5.svg", // 250*250
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"https://bestdori.com/res/icon/band_18.svg", // 18 44*44
	"0",
	"0",
	"https://bestdori.com/res/icon/band_21.svg", // 21 44*44
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"0",
	"https://bestdori.com/res/icon/band_45.svg", // 45
}

var CNTierList = [...]int{
	20, 30, 40, 50, 100, 200, 300, 400, 500, 1000, 1500, 2000, 3000, 4000, 5000, 10000, 15000, 20000, 30000, 40000, 50000, 100000, 150000, 200000, 300000,
}

var JPTierList = [...]int{
	20, 30, 40, 50, 100, 200, 300, 400, 500, 1000, 2000, 5000, 10000, 20000, 30000, 50000, 70000, 100000, 300000,
}

var TWTierList = [...]int{
	20, 30, 40, 50, 100, 500, 1000, 3000, 5000,
}

var ENTierList = [...]int{
	20, 30, 40, 50, 100, 200, 300, 400, 500, 1000, 2000, 2500, 3000, 4000, 5000, 10000,
}

var KRTierList = [...]int{
	100,
}

// Attr 178*179 right upper corner
// Star 174*179 left down corner
// Bandlogo 180*179 left upper corner

type Config struct {
	General struct {
		Token      string `yaml:"token"`
		AdminToken string `yaml:"admin_token"`
	} `yaml:"general"`
	Proxy struct {
		Enabled  bool   `yaml:"enabled" default:"false"`
		Type     string `yaml:"type"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"proxy"`
	Webhook struct {
		Enabled      bool   `yaml:"enabled" default:"false"`
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

type CharacterDetailed struct {
	CharacterType        string   `json:"characterType"`
	CharacterName        []string `json:"characterName"`
	FirstName            []string `json:"firstName"`
	LastName             []string `json:"lastName"`
	Nickname             []string `json:"nickname"`
	BandID               int      `json:"bandId"`
	ColorCode            string   `json:"colorCode"`
	SdAssetBundleName    string   `json:"sdAssetBundleName"`
	DefaultCostumeId     int      `json:"defaultCostumeId"`
	SeasonCostumeListMap any      `json:"seasonCostumeListMap"`
	Ruby                 []string `json:"ruby"`
	Profile              struct {
		CharacterVoice   []string `json:"characterVoice"`
		FavoriteFood     []string `json:"favoriteFood"`
		HatedFood        []string `json:"hatedFood"`
		Hobby            []string `json:"hobby"`
		SelfIntroduction []string `json:"selfIntroduction"`
		School           []string `json:"school"`
		SchoolCls        []string `json:"schoolCls"`
		SchoolYear       []any    `json:"schoolYear"` // It can be int or string, what a dumb design
		Part             string   `json:"part"`
		Birthday         string   `json:"birthday"`
		Constellation    string   `json:"constellation"`
		Height           int      `json:"height"`
	} `json:"profile"`
}

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
	Songs      map[string]RSongsData     `json:"songs"`
	Events     map[string]EventData      `json:"events"`
	Gacha      map[string]RGachaData     `json:"gacha"`
	LoginBonus map[string]LoginBonusData `json:"loginBonus"`
}

type RSongsData struct {
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
	EventType             string             `json:"eventType"`
	EventName             []string           `json:"eventName"`
	AssetBundleName       string             `json:"assetBundleName"`
	BannerAssetBundleName string             `json:"bannerAssetBundleName"`
	StartAt               []string           `json:"startAt"`
	EndAt                 []string           `json:"endAt"`
	Attribute             []Attributes       `json:"attribute"`
	Characters            []EventsCharacters `json:"characters"`
	Members               []any              `json:"members"`
	LimitBreaks           []any              `json:"limitBreaks"`
	RewardCards           []int              `json:"rewardCards"`
}

type Attributes struct {
	Attribute string `json:"attribute"`
	Percent   int    `json:"percent"`
}

type EventsCharacters struct {
	CharacterID any `json:"characterId"`
	Percent     int `json:"percent"`
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
	CharacterID any `json:"characterId"`
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
	Details               []GachaDetailedData `json:"details"`
	Rates                 []GachaRatesData    `json:"rates"`
	PaymentMethods        []PaymentMethods    `json:"paymentMethods"`
	ResourceName          string              `json:"resourceName"`
	BannerAssetBundleName string              `json:"bannerAssetBundleName"`
	GachaName             []string            `json:"gachaName"`
	PublishedAt           []string            `json:"publishedAt"`
	ClosedAt              []string            `json:"closedAt"`
	Description           []string            `json:"description"`
	Annotation            []string            `json:"annotation"`
	GachaPeriod           []string            `json:"gachaPeriod"`
	GachaType             string              `json:"gachaType"`
	Type                  string              `json:"type"`
	NewCards              []int               `json:"newCards"`
	Information           struct {
		Description   []string `json:"description"`
		Term          []string `json:"term"`
		NewMemberInfo []string `json:"newMemberInfo"`
		Notice        []string `json:"notice"`
	} `json:"information"`
}

type DetailsGacha struct {
	RarityIndex int  `json:"rarityIndex"`
	Weight      int  `json:"weight"`
	Pickup      bool `json:"pickup"`
}

type GachaDetailedData map[string]DetailsGacha

type GachaRates struct {
	Rate        float32 `json:"rate"`
	WeightTotal int     `json:"weightTotal"`
}

type GachaRatesData map[string]GachaRates

type PaymentMethods struct {
	GachaID          int    `json:"gachaId"`
	PaymentMethod    string `json:"paymentMethod"`
	Quantity         int    `json:"quantity"`
	PaymentMethodID  int    `json:"paymentMethodId"`
	Count            int    `json:"count"`
	Behavior         string `json:"behavior"`
	Pickup           bool   `json:"pickup"`
	CostItemQuantity int    `json:"costItemQuantity"`
	DiscountType     int    `json:"discountType"`
}

type EventTracker struct {
	Cutoffs []Cutoff `json:"cutoffs"`
	Result  bool     `json:"result"`
}

type Cutoff struct {
	Time int64 `json:"time"`
	Ep   int64 `json:"ep"`
}

type Rates struct {
	Type   string  `json:"type"`
	Server int     `json:"server"`
	Tier   int     `json:"tier"`
	Rate   float64 `json:"rate"`
}

type RatesData []Rates

type Songs struct {
	Tag         string                    `json:"tag"` // `normal` for og, `anime` for covers and `tie-up` for extras
	BandID      int                       `json:"bandId"`
	JacketImage []string                  `json:"jacketImage"`
	MusicTitle  []string                  `json:"musicTitle"`
	PublishedAt []string                  `json:"publishedAt"`
	ClosedAt    []string                  `json:"closedAt"`
	Difficulty  map[string]DifficultyData `json:"difficulty"`
}

type DifficultyData struct {
	PlayLevel   int      `json:"playLevel"`
	PublishedAt []string `json:"publishedAt"`
}

type SongsData map[string]Songs

type SongsDetailed struct {
	BgmID        string                        `json:"bgmId"`
	BgmFile      string                        `json:"bgmFile"`
	Tag          string                        `json:"tag"`
	BandID       int                           `json:"bandId"`
	Achievements []AchievementsData            `json:"achievements"`
	JacketImage  []string                      `json:"jacketImage"`
	Seq          int                           `json:"seq"`
	MusicTitle   []string                      `json:"musicTitle"`
	Ruby         []string                      `json:"ruby"`
	Phonetic     []string                      `json:"phonetic"`
	Lyricist     []string                      `json:"lyricist"`
	Composer     []string                      `json:"composer"`
	Arranger     []string                      `json:"arranger"`
	HowToGet     []string                      `json:"howToGet"`
	PublishedAt  []string                      `json:"publishedAt"`
	ClosedAt     []string                      `json:"closedAt"`
	Description  []string                      `json:"description"`
	Difficulty   map[string]DifficultyDetailed `json:"difficulty"`
	Length       float64                       `json:"length"`
	Notes        map[string]int                `json:"notes"`
	Bpm          map[string]BpmData            `json:"bpm"`
}

type AchievementsData struct {
	MusicID         int    `json:"musicId"`
	AchievementType string `json:"achievementType"`
	RewardType      string `json:"rewardType"`
	Quantity        int    `json:"quantity"`
}

type DifficultyDetailed struct {
	PlayLevel         int                              `json:"playLevel"`
	MultiLiveScoreMap map[string]MultiLiveScoreMapData `json:"multiLiveScoreMap"`
	NotesQuantity     int                              `json:"notesQuantity"`
	ScoreC            int                              `json:"scoreC"`
	ScoreB            int                              `json:"scoreB"`
	ScoreA            int                              `json:"scoreA"`
	ScoreS            int                              `json:"scoreS"`
	ScoreSS           int                              `json:"scoreSS"`
}

type MultiLiveScoreMapData struct {
	MusicID                 int    `json:"musicId"`
	MusicDifficulty         string `json:"musicDifficulty"`
	MusicLiveDifficultyID   int    `json:"musicLiveDifficultyId"`
	ScoreS                  int    `json:"scoreS"`
	ScoreA                  int    `json:"scoreA"`
	ScoreB                  int    `json:"scoreB"`
	ScoreC                  int    `json:"scoreC"`
	MultiLiveDifficultyType string `json:"multiLiveDifficultyType"`
	ScoreSS                 int    `json:"scoreSS"`
	ScoreSSS                int    `json:"scoreSSS"`
}

type BpmData []struct {
	Bpm   int     `json:"bpm"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type ComicData map[string]Comic

type Comic struct {
	AssetBundleName string   `json:"assetBundleName"`
	Title           []string `json:"title"`
	SubTitle        []string `json:"subTitle"`
	PublicStartAt   []any    `json:"publicStartAt"`
	CharacterId     []int    `json:"characterId"`
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
// Now I know the reason, some songs are performed by other combinations of bands, so they have their own IDs.
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

var AcceptedRegions = []string{"jp", "en", "tw", "cn", "kr"}

// These cards are known to have issues with their data, so they are excluded from the bot's responses.
// Maybe they fixed now, but I will leave them here just in case. If you find any other cards with issues, please report them to me.
// Symptom: when send rich message, the telegram api will return 400 Bad Request with message "wrong type of the web page content".
var BadCards = []string{"2309"}

var TelegramAcceptPorts = []int{80, 88, 443, 8443}

var BandWithEmoji = []int{1, 2, 3, 4, 5, 18, 21, 45}

var DifficultyColors = []string{
	"#3076F8", // Easy
	"#188A28", // Normal
	"#D1861A", // Hard
	"#FD2A30", // Expert
	"#F12199", // Special
}

type Item struct {
	ID      int
	Name    string   // 正式名称/主名
	Aliases []string // 别称
}

type searchEntry struct {
	Text string // 用于匹配的字符串（别称或主名）
	Item *Item  // 指向原实体
}

type searchEntries []searchEntry

func (s searchEntries) Len() int {
	return len(s)
}

func (s searchEntries) String(i int) string {
	return s[i].Text
}

type SearchIndex struct {
	entries searchEntries
}

func NewSearchIndex(items []*Item) *SearchIndex {
	var entries searchEntries
	for _, item := range items {
		entries = append(entries, searchEntry{Text: item.Name, Item: item})

		for _, alias := range item.Aliases {
			entries = append(entries, searchEntry{Text: alias, Item: item})
		}
	}
	return &SearchIndex{entries: entries}
}

type SearchResult struct {
	Item         *Item
	MatchedAlias string // 命中哪个别称
	Score        int    // 匹配得分
}

func (idx *SearchIndex) Search(query string) []SearchResult {
	if query == "" {
		return nil
	}

	matches := fuzzy.FindFrom(query, idx.entries)

	var results []SearchResult
	seen := make(map[int]bool) // 用于按 Item.ID 去重

	for _, match := range matches {
		entry := idx.entries[match.Index]
		itemID := entry.Item.ID

		// 如果该 Item 已经有更高分数的别称被命中，则跳过
		if seen[itemID] {
			continue
		}
		seen[itemID] = true

		results = append(results, SearchResult{
			Item:         entry.Item,
			MatchedAlias: entry.Text,
			Score:        match.Score,
		})
	}

	return results
}

var CharaSearch = []*Item{
	{ID: 1, Name: "Kasumi Toyama", Aliases: []string{"とやま かすみ", "토야마 카스미", "户山 香澄", "戶山 香澄", "戸山 香澄", "ksm", "toyama kasumi"}},
	{ID: 2, Name: "Tae Hanazono", Aliases: []string{"はなぞの たえ", "하나조노 타에", "花园 多惠", "花園 多惠", "花園 たえ", "otae", "hanazono tae"}},
	{ID: 3, Name: "Rimi Ushigome", Aliases: []string{"うしごめ りみ", "우시고메 리미", "牛込 里美", "牛込 りみ", "rimirin", "ushigome rimi"}},
	{ID: 4, Name: "Saya Yamabuki", Aliases: []string{"やまぶき さあや", "야마부키 사아야", "山吹 沙绫", "山吹 沙綾", "Saaya", "yamabuki saaya", "yamabuki saya"}},
	{ID: 5, Name: "Arisa Ichigaya", Aliases: []string{"いちがや ありさ", "이치가야 아리사", "市谷 有咲", "市ヶ谷 有咲", "ars", "ichigaya arisa"}},
	{ID: 6, Name: "Ran Mitake", Aliases: []string{"みたけ らん", "미타케 란", "美竹 蘭", "美竹 兰", "ran", "mitake ran"}},
	{ID: 7, Name: "Moca Aoba", Aliases: []string{"あおば もか", "아오바 모카", "青叶 摩卡", "青葉 摩卡", "moca", "aoba moca"}},
	{ID: 8, Name: "Himari Uehara", Aliases: []string{"うえはら ひまり", "우에하라 히마리", "上原 绯玛丽", "上原 ひまり", "上原 緋瑪麗", "hmr", "uehara himari"}},
	{ID: 9, Name: "Tomoe Udagawa", Aliases: []string{"うだがわ ともえ", "우다가와 토모에", "宇田川 巴", "tomoe", "udagawa tomoe"}},
	{ID: 10, Name: "Tsugumi Hazawa", Aliases: []string{"はざわ つぐみ", "하자와 츠구미", "羽泽 鸫", "羽澤 鶇", "羽沢 つぐみ", "tsugumi", "hazawa tsugumi", "tgm"}},
	{ID: 11, Name: "Kokoro Tsurumaki", Aliases: []string{"つるまき こころ", "츠루마키 코코로", "弦卷 心", "弦巻 こころ", "kkr", "tsurumaki kokoro"}},
	{ID: 12, Name: "Kaoru Seta", Aliases: []string{"せた かおる", "세타 카오루", "瀨田 薰", "濑田 薰", "seta kaoru", "kaoru"}},
	{ID: 13, Name: "Hagumi Kitazawa", Aliases: []string{"きたざわ はぐみ", "키타자와 하구미", "北澤 育美", "北泽 育美", "北沢 はぐみ", "kitazawa hagumi", "hgm"}},
	{ID: 14, Name: "Kanon Matsubara", Aliases: []string{"まつばら かのん", "마츠바라 카논", "松原 花音", "kanon", "matsubara kanon"}},
	{ID: 15, Name: "Misaki Okusawa", Aliases: []string{"みっしぇる", "ミッシェル", "米歇爾", "米歇尔", "Michelle", "okusawa misaki", "misaki", "msk"}},
	{ID: 16, Name: "Aya Maruyama", Aliases: []string{"まるやま あや", "마루야마 아야", "丸山 彩", "aya", "maruyama aya"}},
	{ID: 17, Name: "Hina Hikawa", Aliases: []string{"ひかわ ひな", "히카와 히나", "冰川 日菜", "氷川 日菜", "hina", "hikawa hina"}},
	{ID: 18, Name: "Chisato Shirasagi", Aliases: []string{"しらさぎ ちさと", "시라사기 치사토", "白鷺 千聖", "白鹭 千圣", "shirasagi chisato", "chisato", "cst"}},
	{ID: 19, Name: "Maya Yamato", Aliases: []string{"やまと まや", "야마토 마야", "大和 麻弥", "大和 麻彌", "yamato maya", "maya"}},
	{ID: 20, Name: "Eve Wakamiya", Aliases: []string{"わかみや いぶ", "와카미야 이브", "若宮 伊芙", "若宮 イヴ", "若宫 伊芙", "wakamiya eve", "eve"}},
	{ID: 21, Name: "Yukina Minato", Aliases: []string{"みなと ゆきな", "미나토 유키나", "湊 友希那", "湊 友希那", "ykn", "minato yukina"}},
	{ID: 22, Name: "Sayo Hikawa", Aliases: []string{"ひかわ さよ", "히카와 사요", "冰川 紗夜", "冰川 纱夜", "氷川 紗夜", "sayo", "hikawa sayo"}},
	{ID: 23, Name: "Lisa Imai", Aliases: []string{"いまい りさ", "이마이 리사", "今井 莉莎", "今井 リサ", "imai lisa", "lisa", "lisa nee"}},
	{ID: 24, Name: "Ako Udagawa", Aliases: []string{"うだがわ あこ", "우다가와 아코", "宇田川 亞子", "宇田川 亚子", "宇田川 あこ", "ako", "udagawa ako"}},
	{ID: 25, Name: "Rinko Shirokane", Aliases: []string{"しろかね りんこ", "시로카네 린코", "白金 燐子", "shirokane rinko", "Rinko"}},
	{ID: 26, Name: "Mashiro Kurata", Aliases: []string{"くらた ましろ", "쿠라타 마시로", "倉田 真白", "仓田 真白", "倉田 ましろ", "kurata mashiro", "mashiro", "msr"}},
	{ID: 27, Name: "Toko Kirigaya", Aliases: []string{"きりがや とうこ", "키리가야 토우코", "桐谷 透子", "桐ヶ谷 透子", "kirigaya touko", "toko"}},
	{ID: 28, Name: "Nanami Hiromachi", Aliases: []string{"ひろまち ななみ", "히로마치 나나미", "廣町 七深", "广町 七深", "広町 七深", "hiromachi nanami", "nanami", "nnm"}},
	{ID: 29, Name: "Tsukushi Futaba", Aliases: []string{"ふたば つくし", "후타바 츠쿠시", "二葉 筑紫", "二叶 筑紫", "二葉 つくし", "futaba tsukushi", "ftb"}},
	{ID: 30, Name: "Rui Yashio", Aliases: []string{"やしお るい", "야시오 루이", "八潮 瑠唯", "rui", "yashio rui"}},
	{ID: 31, Name: "Rei Wakaba", Aliases: []string{"わかな れい", "와카나 레이", "和奏 瑞依", "LAYER", "和奏 レイ", "wakaba rei", "rei", "レイヤ", "레이야"}},
	{ID: 32, Name: "Rokka Asahi", Aliases: []string{"あさひ ろっか", "아사히 롯카", "朝日 六花", "LOCK", "asahi rokka", "rokka", "ロック", "록"}},
	{ID: 33, Name: "Masuki Sato", Aliases: []string{"さとう ますき", "사토 마스키", "佐藤 ますき", "MASKING", "佐藤 益木", "satou masuki", "masuki", "マスキング", "마스킹"}},
	{ID: 34, Name: "Reona Nyubara", Aliases: []string{"にゅうばら れおな", "뉴바라 레오나", "鳰原 令王那", "PAREO", "nyubara reona", "reona", "パレオ", "파레오"}},
	{ID: 35, Name: "Chiyu Tamade", Aliases: []string{"たまで ちゆ", "타마데 치유", "Chiyu-tan", "珠手 ちゆ", "珠手 知由", "CHU²", "tamade chiyu", "chiyu", "chuchu", "チュチュ", "츄츄"}},
	{ID: 36, Name: "Tomori Takamatsu", Aliases: []string{"たかまつ ともり", "高松 燈", "Takamatsu Tomori", "tomori", "tomorin"}},
	{ID: 37, Name: "Anon Chihaya", Aliases: []string{"ちはや あのん", "千早 愛音", "Chihaya anon", "Anon"}},
	{ID: 38, Name: "Rana Kaname", Aliases: []string{"かなめ らーな", "要 楽奈", "要 樂奈", "要 乐奈", "raana", "kaname rana", "kaname raana"}},
	{ID: 39, Name: "Soyo Nagasaki", Aliases: []string{"ながさき そよ", "長崎 そよ", "長崎 爽世", "长崎 爽世", "nagasaki soyo", "soyo", "soyorin"}},
	{ID: 40, Name: "Taki Shina", Aliases: []string{"しいな たき", "椎名 立希", "shiina taki", "rikki"}},
}
