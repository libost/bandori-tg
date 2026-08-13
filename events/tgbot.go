package events

import (
	"log"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/recent"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("events", eventsCommand))
}

func eventTimingParagraph(startLabel, endLabel, dlangCode string, startAt, endAt int64) gotgbot.RichTextArray {
	return gotgbot.RichTextArray{
		gotgbot.RichTextString(I.GetLocalisedString("events.start_at", dlangCode)),
		gotgbot.RichTextDateTime{
			Text:           gotgbot.RichTextString(startLabel),
			UnixTime:       startAt,
			DateTimeFormat: "DwT",
		},
		gotgbot.RichTextString("\n"),
		gotgbot.RichTextString(I.GetLocalisedString("events.end_at", dlangCode)),
		gotgbot.RichTextDateTime{
			Text:           gotgbot.RichTextString(endLabel),
			UnixTime:       endAt,
			DateTimeFormat: "DwT",
		},
	}
}

func eventsCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		ongoingIDs, err := recent.GetOngoingEventsID()
		if err != nil {
			return err
		}
		// JP events
		eventNameJP := Events[ongoingIDs[0]].EventName[0]
		eventTypeJP := Events[ongoingIDs[0]].EventType
		eventAssetsBundleJP := Events[ongoingIDs[0]].AssetBundleName
		locJP, _ := time.LoadLocation("Asia/Tokyo")
		eStartJP, _ := strconv.Atoi(Events[ongoingIDs[0]].StartAt[0])
		eStartJPTime := time.Unix(int64(eStartJP)/1000, 0).In(locJP).Format("2006-01-02 15:04:05 MST")
		eventStartJP := int64(eStartJP) / 1000
		eEndJP, _ := strconv.Atoi(Events[ongoingIDs[0]].EndAt[0])
		eEndJPTime := time.Unix(int64(eEndJP)/1000, 0).In(locJP).Format("2006-01-02 15:04:05 MST")
		eventEndJP := int64(eEndJP) / 1000
		// EN events
		eventNameEN := Events[ongoingIDs[1]].EventName[1]
		eventTypeEN := Events[ongoingIDs[1]].EventType
		eventAssetsBundleEN := Events[ongoingIDs[1]].AssetBundleName
		locEN, _ := time.LoadLocation("America/New_York")
		eStartEN, _ := strconv.Atoi(Events[ongoingIDs[1]].StartAt[1])
		eStartENTime := time.Unix(int64(eStartEN)/1000, 0).In(locEN).Format("2006-01-02 15:04:05 MST")
		eventStartEN := int64(eStartEN) / 1000
		eEndEN, _ := strconv.Atoi(Events[ongoingIDs[1]].EndAt[1])
		eEndENTime := time.Unix(int64(eEndEN)/1000, 0).In(locEN).Format("2006-01-02 15:04:05 MST")
		eventEndEN := int64(eEndEN) / 1000
		// TW events
		eventNameTW := Events[ongoingIDs[2]].EventName[2]
		eventTypeTW := Events[ongoingIDs[2]].EventType
		eventAssetsBundleTW := Events[ongoingIDs[2]].AssetBundleName
		locTW, _ := time.LoadLocation("Asia/Tokyo")
		eStartTW, _ := strconv.Atoi(Events[ongoingIDs[2]].StartAt[2])
		eStartTWTime := time.Unix(int64(eStartTW)/1000, 0).In(locTW).Format("2006-01-02 15:04:05 MST")
		eventStartTW := int64(eStartTW) / 1000
		eEndTW, _ := strconv.Atoi(Events[ongoingIDs[2]].EndAt[2])
		eEndTWTime := time.Unix(int64(eEndTW)/1000, 0).In(locTW).Format("2006-01-02 15:04:05 MST")
		eventEndTW := int64(eEndTW) / 1000
		// CN events
		eventNameCN := Events[ongoingIDs[3]].EventName[3]
		eventTypeCN := Events[ongoingIDs[3]].EventType
		eventAssetsBundleCN := Events[ongoingIDs[3]].AssetBundleName
		locCN, _ := time.LoadLocation("Asia/Shanghai")
		eStartCN, _ := strconv.Atoi(Events[ongoingIDs[3]].StartAt[3])
		eStartCNTime := time.Unix(int64(eStartCN)/1000, 0).In(locCN).Format("2006-01-02 15:04:05 MST")
		eventStartCN := int64(eStartCN) / 1000
		eEndCN, _ := strconv.Atoi(Events[ongoingIDs[3]].EndAt[3])
		eEndCNTime := time.Unix(int64(eEndCN)/1000, 0).In(locCN).Format("2006-01-02 15:04:05 MST")
		eventEndCN := int64(eEndCN) / 1000
		richMessage := gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.ongoing_events", langCode),
					),
					Size: 4,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(
							"https://bestdori.com/assets/jp/event/" + eventAssetsBundleJP + "/images_rip/banner.png",
						),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(
							eventNameJP + " (" + I.GetLocalisedString("events."+eventTypeJP, langCode) + ")" + " (JP🇯🇵)",
						),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: eventTimingParagraph(
						eStartJPTime,
						eEndJPTime,
						langCode,
						eventStartJP,
						eventEndJP,
					),
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(
							"https://bestdori.com/assets/en/event/" + eventAssetsBundleEN + "/images_rip/banner.png",
						),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(
							eventNameEN + " (" + I.GetLocalisedString("events."+eventTypeEN, langCode) + ")" + " (EN🇺🇸)",
						),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: eventTimingParagraph(
						eStartENTime,
						eEndENTime,
						langCode,
						eventStartEN,
						eventEndEN,
					),
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(
							"https://bestdori.com/assets/tw/event/" + eventAssetsBundleTW + "/images_rip/banner.png",
						),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(
							eventNameTW + " (" + I.GetLocalisedString("events."+eventTypeTW, langCode) + ")" + " (TW🇹🇼)",
						),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: eventTimingParagraph(
						eStartTWTime,
						eEndTWTime,
						langCode,
						eventStartTW,
						eventEndTW,
					),
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(
							"https://bestdori.com/assets/cn/event/" + eventAssetsBundleCN + "/images_rip/banner.png",
						),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(
							eventNameCN + " (" + I.GetLocalisedString("events."+eventTypeCN, langCode) + ")" + " (CN🇨🇳)",
						),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: eventTimingParagraph(
						eStartCNTime,
						eEndCNTime,
						langCode,
						eventStartCN,
						eventEndCN,
					),
				},
			},
		}
		_, err = b.SendRichMessage(ctx.EffectiveUser.Id, richMessage, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		return nil
	} else {
		log.Printf("User %d requested event information with invalid arguments: %v", ctx.EffectiveUser.Id, ctx.Args())
		return nil
	}
}
