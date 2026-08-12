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
		locJP, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return err
		}
		eStartJP, _ := strconv.Atoi(Events[ongoingIDs[0]].StartAt[0])
		eventStartJP := time.UnixMilli(int64(eStartJP)).In(locJP).Format("2006-01-02 15:04:05 MST")
		eEndJP, _ := strconv.Atoi(Events[ongoingIDs[0]].EndAt[0])
		eventEndJP := time.UnixMilli(int64(eEndJP)).In(locJP).Format("2006-01-02 15:04:05 MST")
		// EN events
		eventNameEN := Events[ongoingIDs[1]].EventName[1]
		eventTypeEN := Events[ongoingIDs[1]].EventType
		eventAssetsBundleEN := Events[ongoingIDs[1]].AssetBundleName
		locEN, err := time.LoadLocation("America/New_York")
		if err != nil {
			return err
		}
		eStartEN, _ := strconv.Atoi(Events[ongoingIDs[1]].StartAt[1])
		eventStartEN := time.UnixMilli(int64(eStartEN)).In(locEN).Format("2006-01-02 15:04:05 MST")
		eEndEN, _ := strconv.Atoi(Events[ongoingIDs[1]].EndAt[1])
		eventEndEN := time.UnixMilli(int64(eEndEN)).In(locEN).Format("2006-01-02 15:04:05 MST")
		// TW events
		eventNameTW := Events[ongoingIDs[2]].EventName[2]
		eventTypeTW := Events[ongoingIDs[2]].EventType
		eventAssetsBundleTW := Events[ongoingIDs[2]].AssetBundleName
		locTW, err := time.LoadLocation("Asia/Taipei")
		if err != nil {
			return err
		}
		eStartTW, _ := strconv.Atoi(Events[ongoingIDs[2]].StartAt[2])
		eventStartTW := time.UnixMilli(int64(eStartTW)).In(locTW).Format("2006-01-02 15:04:05 MST")
		eEndTW, _ := strconv.Atoi(Events[ongoingIDs[2]].EndAt[2])
		eventEndTW := time.UnixMilli(int64(eEndTW)).In(locTW).Format("2006-01-02 15:04:05 MST")
		// CN events
		eventNameCN := Events[ongoingIDs[3]].EventName[3]
		eventTypeCN := Events[ongoingIDs[3]].EventType
		eventAssetsBundleCN := Events[ongoingIDs[3]].AssetBundleName
		locCN, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return err
		}
		eStartCN, _ := strconv.Atoi(Events[ongoingIDs[3]].StartAt[3])
		eventStartCN := time.UnixMilli(int64(eStartCN)).In(locCN).Format("2006-01-02 15:04:05 MST")
		eEndCN, _ := strconv.Atoi(Events[ongoingIDs[3]].EndAt[3])
		eventEndCN := time.UnixMilli(int64(eEndCN)).In(locCN).Format("2006-01-02 15:04:05 MST")
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
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.start_at", langCode) + eventStartJP + "\n" +
							I.GetLocalisedString("events.end_at", langCode) + eventEndJP,
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
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.start_at", langCode) + eventStartEN + "\n" +
							I.GetLocalisedString("events.end_at", langCode) + eventEndEN,
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
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.start_at", langCode) + eventStartTW + "\n" +
							I.GetLocalisedString("events.end_at", langCode) + eventEndTW,
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
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.start_at", langCode) + eventStartCN + "\n" +
							I.GetLocalisedString("events.end_at", langCode) + eventEndCN,
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
