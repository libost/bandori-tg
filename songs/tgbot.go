package songs

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/utils"
	"github.com/sahilm/fuzzy"
)

func fuzzySearch(query string, songs []string) []string {
	fuzzyResults := fuzzy.Find(query, songs)
	var results []string
	for _, result := range fuzzyResults {
		results = append(results, result.Str)
	}
	if len(results) > 5 {
		results = results[:5]
	}
	return results
}

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("search", searchCommand))
	dispatcher.AddHandler(handlers.NewCommand("song", SongCommand))
}

func searchCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) < 2 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("songs.search_usage", langCode), nil)
		return nil
	}
	args := ctx.Args()
	args = args[1:]
	var searchStr strings.Builder
	for i, arg := range args {
		searchStr.WriteString(arg)
		if i < len(args)-1 {
			searchStr.WriteString(" ")
		}
	}
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	var songs []string
	switch qlangCode {
	case "jp":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysJP)
	case "en":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysEN)
	case "cn":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysCN)
	case "kr":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysKR)
	case "tw":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysTW)
	default:
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("songs.search_invalid_region", langCode), nil)
		return nil
	}
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("songs.search_results", langCode), searchStr.String(), strings.ToUpper(qlangCode))+strings.Join(songs, "\n"), nil)
	return err
}

func SongCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) < 2 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("songs.song_usage", langCode), nil)
		return nil
	}
	args := ctx.Args()
	args = args[1:]
	var searchStr strings.Builder
	for i, arg := range args {
		searchStr.WriteString(arg)
		if i < len(args)-1 {
			searchStr.WriteString(" ")
		}
	}
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	var songs []string
	switch qlangCode {
	case "jp":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysJP)
	case "en":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysEN)
	case "cn":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysCN)
	case "kr":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysKR)
	case "tw":
		songs = fuzzySearch(searchStr.String(), utils.SongsKeysTW)
	default:
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("songs.search_invalid_region", langCode), nil)
		return nil
	}
	if len(songs) == 0 {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("songs.search_no_results", langCode), searchStr.String(), strings.ToUpper(qlangCode)), nil)
		return nil
	}
	err := sendSongDetails(b, ctx, songs, qlangCode, langCode)
	return err
}

func regionHelper(region string) (int, string) {
	switch region {
	case "jp":
		return 0, "Asia/Tokyo"
	case "en":
		return 1, "UTC"
	case "tw":
		return 2, "Asia/Taipei"
	case "cn":
		return 3, "Asia/Shanghai"
	case "kr":
		return 4, "Asia/Seoul"
	default:
		return 0, "Asia/Tokyo"
	}
}

func sendSongDetails(b *gotgbot.Bot, ctx *ext.Context, songs []string, region, langCode string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	regionIndex, tz := regionHelper(region)
	songMap := utils.ReadSongs()
	var songId string
	for id, song := range songMap {
		if song.MusicTitle[regionIndex] == songs[0] {
			songId = id
			break
		}
	}
	if songId == "" {
		ctx.EffectiveMessage.Reply(b, "No song ID found for \""+songs[0]+"\" ("+region+")", nil)
		return nil
	}
	detailedSong, err := GetDetailedSong(songId)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error fetching song details: "+err.Error(), nil)
		return nil
	}
	jacketURL, _, err := getJacket(songId, detailedSong.JacketImage[0], region)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error fetching song jacket: "+err.Error(), nil)
		return nil
	}
	bgmURL, err := getBgm(songId, region)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error fetching song BGM: "+err.Error(), nil)
		return nil
	}
	bandMap := utils.ReadBand()
	bandStr := fmt.Sprintf("%d", detailedSong.BandID)
	bandName := bandMap[bandStr].BandName[regionIndex]
	bandNameTable := gotgbot.RichTextArray{gotgbot.RichTextString(bandName)}
	if slices.Contains(C.BandWithEmoji, detailedSong.BandID) {
		bandNameTable = gotgbot.RichTextArray{
			gotgbot.RichTextString(bandName + " "),
			gotgbot.RichTextCustomEmoji{
				CustomEmojiId:   fmt.Sprintf("%d", C.BandEmoji[detailedSong.BandID-1]),
				AlternativeText: "😐",
			},
		}
	}
	difficultyText := fmt.Sprintf("Easy: %d \n Normal: %d \n Hard: %d \n Expert: %d", detailedSong.Difficulty["0"].PlayLevel, detailedSong.Difficulty["1"].PlayLevel, detailedSong.Difficulty["2"].PlayLevel, detailedSong.Difficulty["3"].PlayLevel)
	if len(detailedSong.Difficulty) == 5 {
		difficultyText += fmt.Sprintf(" \n Special: %d", detailedSong.Difficulty["4"].PlayLevel)
	}
	cells := [][]gotgbot.RichBlockTableCell{
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.title", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.MusicTitle[regionIndex]),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.type", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs."+detailedSong.Tag, langCode)),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.lyricist", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.Lyricist[regionIndex]),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.composer", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.Composer[regionIndex]),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.arranger", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.Arranger[regionIndex]),
				Align: "right",
			},
		},
	}
	if len(detailedSong.Description) > 0 && detailedSong.Description[regionIndex] != "" {
		cells = append(cells, []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.description", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.Description[regionIndex]),
				Align: "right",
			},
		})
	}
	cells = append(cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString("ID"),
			Align: "left",
		},
		{
			Text:  gotgbot.RichTextString(fmt.Sprintf("%s", songId)),
			Align: "right",
		},
	})
	var countdownTime int64
	publishedAt, err := strconv.Atoi(detailedSong.PublishedAt[regionIndex])
	loc, _ := time.LoadLocation(tz)
	timeStr := time.Unix(int64(publishedAt)/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	now := time.Now().Unix()
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error parsing published at: "+err.Error(), nil)
		return nil
	}
	countdownTime = max(int64(publishedAt)-int64(now)*1000, 0)
	var countdownStr string
	if countdownTime > 0 {
		secs, mins, hrs, days := events.TimeCalc(countdownTime)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("songs.countdown_not_end", langCode), days, hrs, mins, secs)
	} else {
		countdownStr = I.GetLocalisedString("songs.countdown_end", langCode)
	}

	infoCells := [][]gotgbot.RichBlockTableCell{
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.band", langCode)),
				Align: "left",
			},
			{
				Text:  bandNameTable,
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.difficulty", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(difficultyText),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.length", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(fmt.Sprintf("%.2f s", detailedSong.Length)),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.countdown", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(countdownStr),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.published_at", langCode)),
				Align: "left",
			},
			{
				Text: gotgbot.RichTextDateTime{
					Text:           gotgbot.RichTextString(timeStr),
					UnixTime:       int64(publishedAt) / 1000,
					DateTimeFormat: "DwT",
				},
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("songs.how_to_get", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(detailedSong.HowToGet[regionIndex]),
				Align: "right",
			},
		},
	}
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("songs.header", langCode)),
				Size: 3,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(jacketURL),
				},
			},
			gotgbot.InputRichBlockAudio{
				Audio: gotgbot.InputMediaAudio{
					Media:     gotgbot.InputFileByURL(bgmURL),
					Performer: bandName,
					Title:     detailedSong.MusicTitle[regionIndex],
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: cells,
			},
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("songs.info", langCode)),
				Size: 4,
			},
			gotgbot.InputRichBlockTable{
				Cells: infoCells,
			},
		},
	}
	_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
