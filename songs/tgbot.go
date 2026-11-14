package songs

import (
	"context"
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
	"github.com/libost/gotgbot/v2/ext/handlers/filters/callbackquery"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("search", searchCommand))
	dispatcher.AddHandler(handlers.NewCommand("song", SongCommand))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("song_"), songCallbackHandler))
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
	index := utils.NewSearchIndex(utils.SongsKeys)
	songs := index.Search(searchStr.String())
	if len(songs) == 0 {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("songs.search_no_results", langCode), searchStr.String(), strings.ToUpper(qlangCode)), nil)
		return nil
	}
	qidx := utils.IndexHelper(qlangCode)
	var results strings.Builder
	for i, song := range songs {
		if i >= 5 {
			break
		}
		if qidx == 0 {
			if song.Item.Name == "" {
				continue
			}
			fmt.Fprintf(&results, "%d. %s\n", i+1, song.Item.Name)
		} else {
			if song.Item.Aliases[qidx-1] == "" {
				continue
			}
			fmt.Fprintf(&results, "%d. %s\n", i+1, song.Item.Aliases[qidx-1])
		}
	}
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("songs.search_results", langCode), searchStr.String(), strings.ToUpper(qlangCode))+results.String(), nil)
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
	index := utils.NewSearchIndex(utils.SongsKeys)
	songs := index.Search(searchStr.String())

	if len(songs) == 0 {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("songs.search_no_results", langCode), searchStr.String(), strings.ToUpper(qlangCode)), nil)
		return nil
	}
	matchedSong := strconv.Itoa(songs[0].Item.ID)
	err := sendSongDetails(b, ctx, matchedSong, qlangCode, langCode)
	return err
}

func songCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
	data := ctx.CallbackQuery.Data
	parts := strings.Split(data, "_")
	if len(parts) < 3 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Invalid callback data format.",
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data format")
	}
	ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      I.GetLocalisedString("songs.loading", langCode),
		ShowAlert: false,
	})
	songId := parts[1]
	return sendSongDetails(b, ctx, songId, parts[2], langCode)
}

func sendSongDetails(b *gotgbot.Bot, ctx *ext.Context, songId string, region, langCode string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	tz := utils.TzHelper(region)
	regionIndex := utils.IndexHelper(region)

	detailedSong, err := GetDetailedSong(songId)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("generic.error", langCode), err.Error()), &gotgbot.SendMessageOpts{
			ParseMode: "HTML",
		})
		return nil
	}
	jacketURL, _, err := GetJacket(songId, detailedSong.JacketImage[0], region)
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
	if slices.Contains(C.BandWithEmoji[:], detailedSong.BandID) {
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
	publishedAtInt64 := utils.TimeCheck(int64(publishedAt) / 1000)
	loc, _ := time.LoadLocation(tz)
	timeStr := time.Unix(publishedAtInt64, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	now := time.Now().Unix()
	countdownTime = max(publishedAtInt64-now, 0)
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
					UnixTime:       publishedAtInt64,
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
				Text: gotgbot.RichTextString(detailedSong.MusicTitle[regionIndex]),
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
		if utils.IsSpecifiedError(err, "failed to get HTTP URL content") {
			log.Printf("Handling specified error for song %s, Id: %s", detailedSong.MusicTitle[regionIndex], songId)
			httpClient := utils.HttpClient
			httpClient.Timeout = 30 * time.Second
			reqSong, errReqSong := http.NewRequestWithContext(ctxTimeout, http.MethodGet, bgmURL, nil)
			if errReqSong != nil {
				return errReqSong
			}
			respSong, errRespSong := httpClient.Do(reqSong)
			if errRespSong != nil {
				return errRespSong
			}
			defer respSong.Body.Close()
			reqJacket, errReq := http.NewRequestWithContext(ctxTimeout, http.MethodGet, jacketURL, nil)
			if errReq != nil {
				return errReq
			}
			respJacket, errResp := httpClient.Do(reqJacket)
			if errResp != nil {
				return errResp
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
			_, errAgain := b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
				ReplyParameters: &gotgbot.ReplyParameters{
					MessageId: ctx.EffectiveMessage.MessageId,
				},
			})
			if errAgain == nil {
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
