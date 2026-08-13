package events

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/recent"
)

func selectLocaleString(strings []string, index int) string {
	var selected string
	if strings[index] != "" && strings[index] != "null" {
		selected = strings[index]
	} else {
		for _, str := range strings {
			if str != "null" {
				selected = str
				break
			}
		}
	}
	return selected
}
func allEqualPercent[T comparable](slice []T) bool {
	if len(slice) <= 1 {
		return true
	}
	first := slice[0]
	for _, v := range slice[1:] {
		if v != first {
			return false
		}
	}
	return true
}

func SendDetailedEvent(b *gotgbot.Bot, ctx *ext.Context, eventID string, langCode string, qlangCode string) error {
	_, exists := Events[eventID]
	if !exists {
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_not_found", langCode), nil)
		return err
	}
	eventDetailed, err := GetDetailedEvents(eventID)
	if err != nil {
		log.Printf("Error fetching detailed event data for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_fetch_error", langCode), nil)
		return err
	}
	if qlangCode == "" {
		qlangCode = I.QueryLangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	}
	eventType := eventDetailed.EventType
	var qindex int
	switch qlangCode {
	case "jp":
		qindex = 0
	case "en":
		qindex = 1
	case "tw":
		qindex = 2
	case "cn":
		qindex = 3
	case "kr":
		qindex = 4
	default:
		qindex = 0
	}
	eventName := selectLocaleString(eventDetailed.EventName, qindex)
	regionCode := regionCodeFromEvent(eventDetailed, qlangCode)
	var rindex int
	var regionText string
	switch regionCode {
	case "jp":
		rindex = 0
		regionText = I.GetLocalisedString("events.jp", langCode)
	case "en":
		rindex = 1
		regionText = I.GetLocalisedString("events.en", langCode)
	case "tw":
		rindex = 2
		regionText = I.GetLocalisedString("events.tw", langCode)
	case "cn":
		rindex = 3
		regionText = I.GetLocalisedString("events.cn", langCode)
	case "kr":
		rindex = 4
		regionText = I.GetLocalisedString("events.kr", langCode)
	default:
		rindex = 0
		regionText = I.GetLocalisedString("events.jp", langCode)
	}
	bannerUrl := "https://bestdori.com/assets/" + regionCode + "/event/" + eventDetailed.AssetBundleName + "/images_rip/banner.png"
	attribute := eventDetailed.Attributes[0].Attribute
	var aindex int
	switch attribute {
	case "powerful":
		aindex = 0
		attribute = "POWERFUL"
	case "cool":
		aindex = 1
		attribute = "COOL"
	case "happy":
		aindex = 2
		attribute = "HAPPY"
	case "pure":
		aindex = 3
		attribute = "PURE"
	}
	startAt := eventDetailed.StartAt[rindex]
	startAtUnix, _ := strconv.ParseInt(startAt, 10, 64)
	startAtUnix = startAtUnix / 1000
	endAt := eventDetailed.EndAt[rindex]
	endAtUnix, _ := strconv.ParseInt(endAt, 10, 64)
	endAtUnix = endAtUnix / 1000
	loc, _ := time.LoadLocation("UTC")
	startAtTime := time.Unix(startAtUnix, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	endAtTime := time.Unix(endAtUnix, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	var secs, mins, hrs, days int64
	var remainingTimeText string
	if startAtUnix > time.Now().Unix() {
		// The event hasn't started yet
		remainingTime := startAtUnix - time.Now().Unix()
		secs, mins, hrs, days = timeCalc(remainingTime)
		remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.not_started", langCode), days, hrs, mins, secs)
	} else if endAtUnix > time.Now().Unix() && startAtUnix < time.Now().Unix() {
		remainingTime := endAtUnix - time.Now().Unix()
		secs, mins, hrs, days = timeCalc(remainingTime)
		remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.remaining_timeTable", langCode), days, hrs, mins, secs)
	} else {
		// The event has ended
		remainingTimeText = I.GetLocalisedString("events.event_endedTable", langCode)
	}
	charaIDs := make([]string, len(eventDetailed.Characters))
	for i, chara := range eventDetailed.Characters {
		charaIDs[i] = strconv.Itoa(chara.CharacterID)
	}
	percents := make([]int, len(eventDetailed.Characters))
	for i, chara := range eventDetailed.Characters {
		percents[i] = chara.Percent
	}
	allEqual := allEqualPercent(percents)
	var length int
	if allEqual {
		length = len(charaIDs) + 1
	} else {
		length = len(charaIDs) + len(percents)
	}
	charaArray := make([]gotgbot.RichText, length)
	if allEqual {
		for i := 0; i < len(charaIDs); i++ {
			charaIDInt, _ := strconv.Atoi(charaIDs[i])
			charaArray[i] = gotgbot.RichTextCustomEmoji{
				CustomEmojiId:   fmt.Sprint(C.CharaEmoji[(charaIDInt - 1)]),
				AlternativeText: "😐",
			}
		}
		charaArray[len(charaIDs)] = gotgbot.RichTextString("   +" + fmt.Sprint(percents[0]) + "%")
	} else {
		for i := 0; i < length; i = i + 2 {
			charaIDInt, _ := strconv.Atoi(charaIDs[i/2])
			charaArray[i] = gotgbot.RichTextCustomEmoji{
				CustomEmojiId:   fmt.Sprint(C.CharaEmoji[(charaIDInt - 1)]),
				AlternativeText: "😐",
			}
			charaArray[i+1] = gotgbot.RichTextString("   +" + fmt.Sprint(percents[i/2]) + "%\n")
		}
	}

	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(
					I.GetLocalisedString("events.event_details", langCode),
				),
				Size: 3,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(bannerUrl),
				},
				Caption: &gotgbot.RichBlockCaption{
					Text: gotgbot.RichTextString(
						eventName + " (" + I.GetLocalisedString("events."+eventType, langCode) + ")",
					),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: [][]gotgbot.RichBlockTableCell{
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Region", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(regionText),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Title", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(eventName),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Type", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events."+eventType, langCode)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Countdown", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(remainingTimeText),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.StartAt", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextDateTime{
								Text:           gotgbot.RichTextString(startAtTime),
								UnixTime:       startAtUnix,
								DateTimeFormat: "DwT",
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.EndAt", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextDateTime{
								Text:           gotgbot.RichTextString(endAtTime),
								UnixTime:       endAtUnix,
								DateTimeFormat: "DwT",
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Attribute", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextArray{
								gotgbot.RichTextString(attribute + " "),
								gotgbot.RichTextCustomEmoji{
									CustomEmojiId:   fmt.Sprint(C.AttributeEmoji[aindex]),
									AlternativeText: "😐",
								},
								gotgbot.RichTextString("   +" + fmt.Sprint(eventDetailed.Attributes[0].Percent) + "%"),
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.Characters", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextArray(charaArray),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.EventID", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(eventID),
							Align: "right",
						},
					},
				},
			},
		},
	}
	_, err = b.SendRichMessage(ctx.EffectiveUser.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err

}

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

func timeCalc(time int64) (int64, int64, int64, int64) {
	if time >= 60 {
		minutes := int64(time / 60)
		seconds := time % 60
		if minutes >= 60 {
			hours := minutes / 60
			minutes = int64(minutes % 60)
			if hours >= 24 {
				days := hours / 24
				hours = int64(hours % 24)
				return seconds, minutes, hours, days
			}
			return seconds, minutes, hours, 0
		}
		return seconds, minutes, 0, 0
	}
	return time, 0, 0, 0
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
		remainingTimeJP := eventEndJP - time.Now().Unix()
		var remainingTimeTextJP string
		if remainingTimeJP <= 0 {
			remainingTimeTextJP = I.GetLocalisedString("events.event_ended", langCode)
		} else {
			secs, mins, hrs, days := timeCalc(remainingTimeJP)
			remainingTimeTextJP = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
		}

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
		remainingTimeEN := eventEndEN - time.Now().Unix()
		var remainingTimeTextEN string
		if remainingTimeEN <= 0 {
			remainingTimeTextEN = I.GetLocalisedString("events.event_ended", langCode)
		} else {
			secs, mins, hrs, days := timeCalc(remainingTimeEN)
			remainingTimeTextEN = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
		}

		// TW events
		eventNameTW := Events[ongoingIDs[2]].EventName[2]
		eventTypeTW := Events[ongoingIDs[2]].EventType
		eventAssetsBundleTW := Events[ongoingIDs[2]].AssetBundleName
		locTW, _ := time.LoadLocation("Asia/Taipei")
		eStartTW, _ := strconv.Atoi(Events[ongoingIDs[2]].StartAt[2])
		eStartTWTime := time.Unix(int64(eStartTW)/1000, 0).In(locTW).Format("2006-01-02 15:04:05 MST")
		eventStartTW := int64(eStartTW) / 1000
		eEndTW, _ := strconv.Atoi(Events[ongoingIDs[2]].EndAt[2])
		eEndTWTime := time.Unix(int64(eEndTW)/1000, 0).In(locTW).Format("2006-01-02 15:04:05 MST")
		eventEndTW := int64(eEndTW) / 1000
		remainingTimeTW := eventEndTW - time.Now().Unix()
		var remainingTimeTextTW string
		if remainingTimeTW <= 0 {
			remainingTimeTextTW = I.GetLocalisedString("events.event_ended", langCode)
		} else {
			secs, mins, hrs, days := timeCalc(remainingTimeTW)
			remainingTimeTextTW = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
		}

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
		remainingTimeCN := eventEndCN - time.Now().Unix()
		var remainingTimeTextCN string
		if remainingTimeCN <= 0 {
			remainingTimeTextCN = I.GetLocalisedString("events.event_ended", langCode)
		} else {
			secs, mins, hrs, days := timeCalc(remainingTimeCN)
			remainingTimeTextCN = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
		}

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
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextString(remainingTimeTextJP),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextUrl{
						Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_detailsClick", langCode)),
						Url:  fmt.Sprintf("https://t.me/%s?start=events_%s_jp", b.Username, ongoingIDs[0]),
					},
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
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextString(remainingTimeTextEN),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextUrl{
						Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_detailsClick", langCode)),
						Url:  fmt.Sprintf("https://t.me/%s?start=events_%s_en", b.Username, ongoingIDs[1]),
					},
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
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextString(remainingTimeTextTW),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextUrl{
						Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_detailsClick", langCode)),
						Url:  fmt.Sprintf("https://t.me/%s?start=events_%s_tw", b.Username, ongoingIDs[2]),
					},
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
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextString(remainingTimeTextCN),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextUrl{
						Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_detailsClick", langCode)),
						Url:  fmt.Sprintf("https://t.me/%s?start=events_%s_cn", b.Username, ongoingIDs[3]),
					},
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
		param := ctx.Args()[1]
		return SendDetailedEvent(b, ctx, param, langCode, "")
	}
}
