package utils

import "github.com/sahilm/fuzzy"

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
	{ID: 35, Name: "Chiyu Tamade", Aliases: []string{"たまで ちゆ", "타마데 치유", "Chiyu-tan", "珠手 ちゆ", "珠手 知由", "CHU²", "tamade chiyu", "chiyu", "chuchu", "chu2", "チュチュ", "츄츄"}},
	{ID: 36, Name: "Tomori Takamatsu", Aliases: []string{"たかまつ ともり", "高松 燈", "Takamatsu Tomori", "tomori", "tomorin"}},
	{ID: 37, Name: "Anon Chihaya", Aliases: []string{"ちはや あのん", "千早 愛音", "Chihaya anon", "Anon"}},
	{ID: 38, Name: "Rana Kaname", Aliases: []string{"かなめ らーな", "要 楽奈", "要 樂奈", "要 乐奈", "raana", "kaname rana", "kaname raana"}},
	{ID: 39, Name: "Soyo Nagasaki", Aliases: []string{"ながさき そよ", "長崎 そよ", "長崎 爽世", "长崎 爽世", "nagasaki soyo", "soyo", "soyorin"}},
	{ID: 40, Name: "Taki Shina", Aliases: []string{"しいな たき", "椎名 立希", "shiina taki", "rikki"}},
}
