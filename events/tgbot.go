package events

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	C "github.com/libost/bandori-tg/constants"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/recent"
	"github.com/libost/bandori-tg/utils"
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

func tzHelper(code string) string {
	switch code {
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
		return "UTC"
	}
}

func SendDetailedEvent(b *gotgbot.Bot, ctx *ext.Context, eventID string, langCode string, qlangCode string) error {
	eventMap := utils.ReadEvents()
	_, exists := eventMap[eventID]
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
		qlangCode = I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
	}
	eventType := eventDetailed.EventType
	qindex := indexHelper(qlangCode)
	eventName := selectLocaleString(eventDetailed.EventName, qindex)
	regionCode := regionCodeFromEvent(eventDetailed, qlangCode)
	if regionCode != qlangCode {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("events.region_no_data", langCode), qlangCode, regionCode), nil)
	}
	rindex := indexHelper(regionCode)
	bannerUrl := "https://bestdori.com/assets/" + regionCode + "/event/" + eventDetailed.AssetBundleName + "/images_rip/banner.png"
	attribute := eventDetailed.Attributes[0].Attribute
	aindex := indexHelper(attribute)
	attribute = strings.ToUpper(attribute)
	startAt := eventDetailed.StartAt[rindex]
	startAtUnix, _ := strconv.ParseInt(startAt, 10, 64)
	startAtUnix = startAtUnix / 1000
	endAt := eventDetailed.EndAt[rindex]
	endAtUnix, _ := strconv.ParseInt(endAt, 10, 64)
	endAtUnix = endAtUnix / 1000
	var tz string
	tz = tzHelper(regionCode)
	loc, _ := time.LoadLocation(tz)
	startAtTime := time.Unix(startAtUnix, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	endAtTime := time.Unix(endAtUnix, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	var secs, mins, hrs, days int64
	var remainingTimeText string
	now := time.Now().Unix()
	if startAtUnix > now {
		// The event hasn't started yet
		remainingTime := startAtUnix - now
		secs, mins, hrs, days = timeCalc(remainingTime)
		remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.not_started", langCode), days, hrs, mins, secs)
	} else if endAtUnix > now && startAtUnix < now {
		remainingTime := endAtUnix - now
		secs, mins, hrs, days = timeCalc(remainingTime)
		remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
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
		for i := range charaIDs {
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
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events."+regionCode, langCode)),
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
	_, err = b.SendRichMessage(ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
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
	stopAction := make(chan struct{})
	var stopActionOnce sync.Once
	stopActionLoop := func() {
		stopActionOnce.Do(func() {
			close(stopAction)
		})
	}
	go func() {
		_, _ = b.SendChatAction(ctx.EffectiveUser.Id, "typing", nil)
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = b.SendChatAction(ctx.EffectiveUser.Id, "typing", nil)
			case <-stopAction:
				return
			}
		}
	}()
	defer stopActionLoop()
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		ongoingIDs, err := recent.GetOngoingEventsID()
		if err != nil {
			return err
		}
		richMessage := gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(
						I.GetLocalisedString("events.ongoing_events", langCode),
					),
					Size: 3,
				},
			},
		}

		// 地区信息：索引、名称、时区、区域代码、emoji
		regions := []struct {
			idx    int
			name   string
			region string
			tz     string
			emoji  string
		}{
			{0, "JP", "jp", "Asia/Tokyo", "🇯🇵"},
			{1, "EN", "en", "UTC", "🇺🇸"},
			{2, "TW", "tw", "Asia/Taipei", "🇹🇼"},
			{3, "CN", "cn", "Asia/Shanghai", "🇨🇳"},
		}

		eventMap := utils.ReadEvents()
		for _, r := range regions {
			event := eventMap[ongoingIDs[r.idx]]
			loc, _ := time.LoadLocation(r.tz)
			eStart, _ := strconv.Atoi(event.StartAt[r.idx])
			eEnd, _ := strconv.Atoi(event.EndAt[r.idx])

			eventStart := int64(eStart) / 1000
			eventEnd := int64(eEnd) / 1000
			remainingTime := eventEnd - time.Now().Unix()
			startRemainingTime := eventStart - time.Now().Unix()

			var remainingTimeText string
			if remainingTime <= 0 {
				remainingTimeText = I.GetLocalisedString("events.event_ended", langCode)
			} else if startRemainingTime > 0 {
				secs, mins, hrs, days := timeCalc(startRemainingTime)
				remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.not_started", langCode), days, hrs, mins, secs)
			} else {
				secs, mins, hrs, days := timeCalc(remainingTime)
				remainingTimeText = fmt.Sprintf(I.GetLocalisedString("events.remaining_time", langCode), days, hrs, mins, secs)
			}

			eStartTime := time.Unix(int64(eStart)/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
			eEndTime := time.Unix(int64(eEnd)/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")

			richMessage.Blocks = append(richMessage.Blocks,
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(
							"https://bestdori.com/assets/" + r.region + "/event/" + event.AssetBundleName + "/images_rip/banner.png",
						),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(
							event.EventName[r.idx] + " (" + I.GetLocalisedString("events."+event.EventType, langCode) + ")" + " (" + r.name + r.emoji + ")",
						),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextString(remainingTimeText),
					},
				},
				gotgbot.InputRichBlockParagraph{
					Text: eventTimingParagraph(
						eStartTime,
						eEndTime,
						langCode,
						eventStart,
						eventEnd,
					),
				},
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextUrl{
						Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_detailsClick", langCode)),
						Url:  fmt.Sprintf("https://t.me/%s?start=events_%s_%s", b.Username, ongoingIDs[r.idx], r.region),
					},
				},
			)
		}
		_, err = b.SendRichMessage(ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		log.Printf("Sent ongoing events to user %d", ctx.EffectiveUser.Id)
		return err
	} else {
		param := ctx.Args()[1]
		qlangCode := ""
		if len(ctx.Args()) > 2 {
			if slices.Contains(C.AcceptedRegions, ctx.Args()[2]) {
				qlangCode = ctx.Args()[2]
			} else {
				_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.invalid_region", langCode), nil)
				return err
			}
		}
		return SendDetailedEvent(b, ctx, param, langCode, qlangCode)
	}
}
