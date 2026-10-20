package gacha

import (
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
	"fmt"
	"strconv"
	"time"

	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
)

func indexHelper(code string) int {
	switch code {
	case "jp", "powerful":
		return 0
	case "en", "cool":
		return 1
	case "tw", "happy":
		return 2
	case "cn", "pure":
		return 3
	case "kr":
		return 4
	default:
		return 0
	}
}

func AddHandlers(dispatcher *ext.Dispatcher) {
	//Remember to register this command after the handler is implemented
	dispatcher.AddHandler(handlers.NewCommand("gacha", gachaHandler))
}

func gachaHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	//TODO: Implement gacha command logic here
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.missing_argument", langCode), nil)
		return err
	}

	_, ok := GetGacha(ctx.Args()[1])
	if ok != nil {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.invalid_argument", langCode), nil)
		return err
	}
	gachaDetailed, err := GetGachaDetails(ctx.Args()[1])
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}

	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)

	qindex := indexHelper(regionCode)

	now := time.Now().Unix() * 1000 // Convert to milliseconds
	var countdownStr string
	publishedAt, parseErr := strconv.ParseInt(gachaDetailed.PublishedAt[qindex], 10, 64)
	if parseErr != nil {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	endAt, parseErr := strconv.ParseInt(gachaDetailed.ClosedAt[qindex], 10, 64)
	if parseErr != nil {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	if publishedAt > now {
		countdown := publishedAt - now
		secs, mins, hrs, days := events.TimeCalc(countdown)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_published", langCode), days, hrs, mins, secs)
	} else if endAt > now {
		countdown := endAt - now
		secs, mins, hrs, days := events.TimeCalc(countdown)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_closed", langCode), days, hrs, mins, secs)
	} else {
		countdownStr = I.GetLocalisedString("gacha.countdown_ended", langCode)
	}

	startTimeStr := time.Unix(publishedAt/1000, 0).Format("2006-01-02 15:04:05")
	endTimeStr := time.Unix(endAt/1000, 0).Format("2006-01-02 15:04:05")

	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("gacha.title", langCode)),
				Size: 3,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL("https://bestdori.com/assets/" + regionCode + "homebanner_rip/" + gachaDetailed.BannerAssetBundleName + ".png"),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: [][]gotgbot.RichBlockTableCell{
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.title", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(gachaDetailed.GachaName[qindex]),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.type", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha."+string(gachaDetailed.GachaType[qindex]), langCode)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.countdown", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(countdownStr),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.start_time", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextDateTime{
								Text:           gotgbot.RichTextString(startTimeStr),
								UnixTime:       publishedAt / 1000,
								DateTimeFormat: "DwT",
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.end_time", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextDateTime{
								Text:           gotgbot.RichTextString(endTimeStr),
								UnixTime:       endAt / 1000,
								DateTimeFormat: "DwT",
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.description", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(gachaDetailed.Description[qindex]),
							Align: "right",
						},
					},
				},
			},
		},
	}

	_, err = b.SendRichMessage(ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
