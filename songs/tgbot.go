package songs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
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
	stopAction := make(chan struct{})
	var stopActionOnce sync.Once
	stopActionLoop := func() {
		stopActionOnce.Do(func() {
			close(stopAction)
		})
	}
	go func() {
		_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
			case <-stopAction:
				return
			}
		}
	}()
	defer stopActionLoop()
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
	if err != nil {
		if ctx.EffectiveChat.Type == "private" {
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
				ParseMode: "HTML",
			})
		} else {
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
				EphemeralMessageParameters: &gotgbot.EphemeralMessageParameters{
					ReceiverUserId: ctx.EffectiveUser.Id,
				},
				ParseMode: "HTML",
			})
		}
	}
	return err
}

func SongCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	stopAction := make(chan struct{})
	var stopActionOnce sync.Once
	stopActionLoop := func() {
		stopActionOnce.Do(func() {
			close(stopAction)
		})
	}
	go func() {
		_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = b.SendChatAction(ctx.EffectiveChat.Id, "typing", nil)
			case <-stopAction:
				return
			}
		}
	}()
	defer stopActionLoop()
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
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return nil
	}
	jacketURL, _, err := getJacket(songId, detailedSong.JacketImage[0], region)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), "Error fetching song jacket: "+err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return nil
	}
	bgmURL, err := getBgm(songId, region)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), "Error fetching song BGM: "+err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
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
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), "Error parsing published at: "+err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return nil
	}
	loc, _ := time.LoadLocation(tz)
	timeStr := time.Unix(int64(publishedAt)/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	now := time.Now().Unix()
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
	if err != nil {
		if isSpecifiedError(err) {
			log.Printf("Handling specified error for song %s", detailedSong.MusicTitle[regionIndex])
			httpClient := utils.HttpClient
			httpClient.Timeout = 30 * time.Second
			reqSong, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, bgmURL, nil)
			if err != nil {
				return err
			}
			respSong, err := httpClient.Do(reqSong)
			if err != nil {
				return err
			}
			defer respSong.Body.Close()
			reqJacket, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, jacketURL, nil)
			if err != nil {
				return err
			}
			respJacket, err := httpClient.Do(reqJacket)
			if err != nil {
				return err
			}
			defer respJacket.Body.Close()
			richMessage.Blocks[1] = gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("jacket.png", respJacket.Body),
				},
			}
			richMessage.Blocks[2] = gotgbot.InputRichBlockAudio{
				Audio: gotgbot.InputMediaAudio{
					Media:     gotgbot.InputFileByReader(fmt.Sprintf("%s.mp3", detailedSong.MusicTitle[regionIndex]), respSong.Body),
					Performer: bandName,
					Title:     detailedSong.MusicTitle[regionIndex],
				},
			}
			_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId: ctx.EffectiveMessage.MessageId,
				},
			})
			if err == nil {
				log.Printf("Successfully handled malformed song %s query", detailedSong.MusicTitle[regionIndex])
				return nil
			} else {
				log.Printf("Failed to send rich message for song %s: %v", detailedSong.MusicTitle[regionIndex], err)
			}
		}
		if ctx.EffectiveChat.Type == "private" {
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
				ParseMode: "HTML",
			})
		} else {
			ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
				EphemeralMessageParameters: &gotgbot.EphemeralMessageParameters{
					ReceiverUserId: ctx.EffectiveUser.Id,
				},
				ParseMode: "HTML",
			})
		}
	}
	return err
}

func isSpecifiedError(err error) bool {
	if err == nil {
		return false
	}
	if tgErr, ok := errors.AsType[*gotgbot.TelegramError](err); ok {
		if tgErr.Code == 400 &&
			tgErr.Method == "sendRichMessage" &&
			strings.Contains(tgErr.Description, "failed to get HTTP URL content") {
			return true
		}
	}
	return false
}
