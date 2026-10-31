package comic

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("comic", comicCommand))
}

func comicCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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
	args := ctx.Args()
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	qidx := indexHelper(regionCode)
	if len(args) == 1 {
		comicKeys := utils.ReadComicKeys()
		if len(comicKeys) == 0 {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("comic.no_comics_available", langCode), nil)
			return fmt.Errorf("no comics available")
		}
		for range comicKeys {
			comicId := comicKeys[rand.Intn(len(comicKeys))]
			comicData, err := GetComicData(comicId)
			if err != nil {
				return err
			}
			if comicData.Title[qidx] != "" {
				err = sendComic(b, ctx, ctxTimeout, comicData, comicId, langCode, regionCode)
				if err != nil {
					return err
				}
				return err
			}
		}
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("comic.no_comics_available_in_specified_region", langCode), nil)
		log.Printf("Comic: No comics available in the specified region: %s", regionCode)
		return err
	}
	comicId := args[1]
	comicData, err := GetComicData(comicId)
	if err != nil {
		return err
	}
	comicTitle := comicData.Title[qidx]
	if comicTitle == "" {
		for i, title := range comicData.Title {
			if title != "" {
				comicTitle = title
				qidx = i
				break
			}
		}
		_, err = ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("comic.not_in_specified_region", langCode), regionCode, reverseIndexHelper(qidx)), nil)
		if err != nil {
			log.Printf("Comic: Failed to send reply message: %v", err)
			return err
		}
	}

	regionCode = reverseIndexHelper(qidx)

	err = sendComic(b, ctx, ctxTimeout, comicData, comicId, langCode, regionCode)
	return err
}

func sendComic(b *gotgbot.Bot, ctx *ext.Context, ctxTimeout context.Context, comic C.Comic, comicId string, langCode string, regionCode string) error {
	url, err := getComicUrl(comic, regionCode)
	if err != nil {
		return err
	}
	comicType := "singleframe"
	if strings.Contains(comic.AssetBundleName, "fourframe") {
		comicType = "fourframe"
	}
	characters := make([]gotgbot.RichText, len(comic.CharacterId))
	for i, charId := range comic.CharacterId {
		characters[i] = gotgbot.RichTextCustomEmoji{
			CustomEmojiId:   fmt.Sprintf("%d", C.CharaEmoji[charId-1]),
			AlternativeText: "😐",
		}
	}

	idx := indexHelper(regionCode)
	cells := [][]gotgbot.RichBlockTableCell{
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.title", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(comic.Title[idx]),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.subtitle", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(comic.SubTitle[idx]),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.type", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString(fmt.Sprintf("comic.type_%s", comicType), langCode)),
				Align: "right",
			},
		},
		{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.characters", langCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextArray(characters),
				Align: "right",
			},
		},
	}
	publicStartAt := convertAnyToInt(comic.PublicStartAt[idx])
	if publicStartAt != 1 {
		timestamp := int64(publicStartAt) / 1000
		timezone := tzHelper(regionCode)
		loc, _ := time.LoadLocation(timezone)
		timeStr := time.Unix(timestamp, 0).In(loc).Format("2006-01-02 15:04:05 MST")
		cells = append(cells, []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.public_start_at", langCode)),
				Align: "left",
			},
			{
				Text: gotgbot.RichTextDateTime{
					Text:           gotgbot.RichTextString(timeStr),
					UnixTime:       timestamp,
					DateTimeFormat: "DwT",
				},
				Align: "right",
			},
		})
	}
	cells = append(cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("comic.id", langCode)),
			Align: "left",
		},
		{
			Text:  gotgbot.RichTextString(comicId),
			Align: "right",
		},
	})
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(comic.Title[idx]),
				Size: 5,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(url),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: cells,
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

func tzHelper(region string) string {
	switch region {
	case "jp":
		return "Asia/Tokyo"
	case "en":
		return "UTC"
	case "tw":
		return "Asia/Taipei"
	case "cn":
		return "Asia/Shanghai"
	case "kr":
		return "Asia/Seoul"
	default:
		return "Asia/Tokyo"
	}
}
