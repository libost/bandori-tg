package events

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
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

func normalizeCharacterID(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
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

func getTimeHelper(eventDetailed C.EventDetailed, field string, index int, tz *time.Location) (int64, string) {
	if field == "start" {
		timeR := eventDetailed.StartAt[index]
		timeInt, _ := strconv.ParseInt(timeR, 10, 64)
		timeInt = timeInt / 1000
		timeStr := time.Unix(timeInt, 0).In(tz).Format("2006-01-02 15:04:05 MST")
		return timeInt, timeStr
	}
	if field == "end" {
		timeR := eventDetailed.EndAt[index]
		timeInt, _ := strconv.ParseInt(timeR, 10, 64)
		timeInt = timeInt / 1000
		timeStr := time.Unix(timeInt, 0).In(tz).Format("2006-01-02 15:04:05 MST")
		return timeInt, timeStr
	}
	return 0, ""
}

func SendDetailedEvent(b *gotgbot.Bot, ctx *ext.Context, eventID string, langCode string, qlangCode string) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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

	var tz string
	tz = tzHelper(regionCode)
	loc, _ := time.LoadLocation(tz)
	startAtUnix, startAtTime := getTimeHelper(eventDetailed, "start", rindex, loc)
	endAtUnix, endAtTime := getTimeHelper(eventDetailed, "end", rindex, loc)

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
	percents := make([]int, len(eventDetailed.Characters))
	for i, chara := range eventDetailed.Characters {
		charaIDs[i] = normalizeCharacterID(chara.CharacterID)
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

	type tableCell struct {
		Type string
		Str  gotgbot.RichText
	}

	table := []tableCell{
		{

			Type: I.GetLocalisedString("events.Region", langCode),
			Str:  gotgbot.RichTextString(I.GetLocalisedString("events."+regionCode, langCode)),
		},
		{
			Type: I.GetLocalisedString("events.Title", langCode),
			Str:  gotgbot.RichTextString(eventName),
		},
		{
			Type: I.GetLocalisedString("events.Type", langCode),
			Str:  gotgbot.RichTextString(I.GetLocalisedString("events."+eventType, langCode)),
		},
		{
			Type: I.GetLocalisedString("events.Countdown", langCode),
			Str:  gotgbot.RichTextString(remainingTimeText),
		},
		{
			Type: I.GetLocalisedString("events.StartAt", langCode),
			Str: &gotgbot.RichTextDateTime{
				Text:           gotgbot.RichTextString(startAtTime),
				UnixTime:       startAtUnix,
				DateTimeFormat: "DwT",
			},
		},
		{
			Type: I.GetLocalisedString("events.EndAt", langCode),
			Str: &gotgbot.RichTextDateTime{
				Text:           gotgbot.RichTextString(endAtTime),
				UnixTime:       endAtUnix,
				DateTimeFormat: "DwT",
			},
		},
		{
			Type: I.GetLocalisedString("events.Attribute", langCode),
			Str: &gotgbot.RichTextArray{
				gotgbot.RichTextString(attribute + " "),
				gotgbot.RichTextCustomEmoji{
					CustomEmojiId:   fmt.Sprint(C.AttributeEmoji[aindex]),
					AlternativeText: "😐",
				},
				gotgbot.RichTextString("   +" + fmt.Sprint(eventDetailed.Attributes[0].Percent) + "%"),
			},
		},
		{
			Type: I.GetLocalisedString("events.Characters", langCode),
			Str:  gotgbot.RichTextArray(charaArray),
		},
		{
			Type: I.GetLocalisedString("events.EventID", langCode),
			Str:  gotgbot.RichTextString(eventID),
		},
	}

	richCells := make([][]gotgbot.RichBlockTableCell, len(table))
	for i, cell := range table {
		richCells[i] = []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(cell.Type),
				Align: "left",
			},
			{
				Text:  cell.Str,
				Align: "right",
			},
		}
	}

	img, err := generateEventMembers(eventDetailed)
	if err != nil {
		log.Printf("Error generating event members image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", langCode), nil)
		return err
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Printf("Error encoding event members image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", langCode), nil)
		return err
	}
	img2, err := generateEventBonusCards(eventDetailed)
	var buf2 bytes.Buffer
	err = png.Encode(&buf2, img2)
	if err != nil {
		log.Printf("Error encoding event bonus cards image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", langCode), nil)
		return err
	}
	fileID, err := utils.GetImageIDWorkAround(b, buf)
	if err != nil {
		log.Printf("Error getting image ID for event members image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", langCode), nil)
		return err
	}
	fileID2, err := utils.GetImageIDWorkAround(b, buf2)
	if err != nil {
		log.Printf("Error getting image ID for event bonus cards image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", langCode), nil)
		return err
	}
	photoBlock := gotgbot.InputRichBlockPhoto{
		Photo: gotgbot.InputMediaPhoto{
			Media: gotgbot.InputFileByID(fileID),
		},
	}
	photoBlock2 := gotgbot.InputRichBlockPhoto{
		Photo: gotgbot.InputMediaPhoto{
			Media: gotgbot.InputFileByID(fileID2),
		},
	}

	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(
					I.GetLocalisedString("events.event_details", langCode),
				),
				Size: 2,
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
				Cells: richCells,
			},
			gotgbot.InputRichBlockDivider{},
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_members", langCode)),
				Size: 3,
			},
			gotgbot.InputRichBlockDetails{
				Summary: gotgbot.RichTextString(""),
				Blocks: []gotgbot.InputRichBlock{
					photoBlock,
				},
			},
			gotgbot.InputRichBlockDivider{},
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString("events.event_bonus_cards", langCode)),
				Size: 3,
			},
			gotgbot.InputRichBlockDetails{
				Summary: gotgbot.RichTextString(""),
				Blocks: []gotgbot.InputRichBlock{
					photoBlock2,
				},
			},
		},
	}

	_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		log.Printf("Request timed out while sending event details to user %d", ctx.EffectiveUser.Id)
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.request_timeout", langCode), nil)
		return err
	}
	if err != nil {
		return err
	}
	return err

}

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("events", eventsCommand))
	dispatcher.AddHandler(handlers.NewCommand("cutoff", FsxCommand))
	dispatcher.AddHandler(handlers.NewCommand("fsx", FsxCommand))
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

func FsxCommand(b *gotgbot.Bot, ctx *ext.Context) error {
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
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.fsx_usage", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
		return nil
	}
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")

	tier := 0
	if len(ctx.Args()) > 2 {
		t, err := strconv.Atoi(ctx.Args()[2])
		if err != nil {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.fsx_usage", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
			return nil
		}
		tier = t
	}
	if len(ctx.Args()) > 3 {
		region := ctx.Args()[3]
		if !slices.Contains(C.AcceptedRegions, region) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.invalid_region", langCode), nil)
			return nil
		}
		qlangCode = region
	}
	eventID := ctx.Args()[1]
	regionCode := indexHelper(qlangCode)
	img, last, predicted, latestTime, speed, err := getEventTracker(qlangCode, eventID, tier)
	if err != nil {
		if errors.Is(err, C.ErrNoSuchEvent) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_not_found", langCode), nil)
			return nil
		}
		if errors.Is(err, C.ErrNoCutoffData) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.no_cutoff_data", langCode), nil)
			return nil
		}
	}
	// Do something with the generated image, e.g., send it to the user
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Printf("Error encoding event tracker image for event ID %s: %v", eventID, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.event_image_error", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)), nil)
		return err
	}
	var status string
	events := utils.ReadEvents()
	event, _ := events[eventID]
	now := time.Now().Unix() * 1000
	startAt, err := strconv.ParseInt(event.StartAt[regionCode], 10, 64)
	if err != nil {
		return err
	}
	endAt, err := strconv.ParseInt(event.EndAt[regionCode], 10, 64)
	if err != nil {
		return err
	}
	if now > endAt {
		status = I.GetLocalisedString("events.event_ended", langCode)
	} else if now > startAt {
		percent := int64(float64(now-startAt) / float64(endAt-startAt) * 100)
		status = fmt.Sprintf(I.GetLocalisedString("events.complete_percent", langCode), percent)
	}
	predictedStr := fmt.Sprintf("%d", predicted)
	if predicted == 0 {
		predictedStr = I.GetLocalisedString("events.not_enough_data", langCode)
	}
	tz := tzHelper(qlangCode)
	loc, _ := time.LoadLocation(tz)
	lastTimeStr := time.Unix(latestTime/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")

	fileId, err := utils.GetImageIDWorkAround(b, buf)
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByID(fileId),
				},
			},
			gotgbot.InputRichBlockTable{
				Cells: [][]gotgbot.RichBlockTableCell{
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.status", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(status),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.latest_cutoff", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(fmt.Sprintf("%d", last)),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.predicted_cutoff", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(predictedStr),
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.updated_at", langCode)),
							Align: "left",
						},
						{
							Text: gotgbot.RichTextDateTime{
								Text:           gotgbot.RichTextString(lastTimeStr),
								UnixTime:       latestTime / 1000,
								DateTimeFormat: "DwT",
							},
							Align: "right",
						},
					},
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("events.speed", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(fmt.Sprintf("%d pt/h", speed)),
							Align: "right",
						},
					},
				},
			},
			gotgbot.InputRichBlockDivider{},
			gotgbot.InputRichBlockBlockQuotation{
				Blocks: []gotgbot.InputRichBlock{
					gotgbot.InputRichBlockParagraph{
						Text: gotgbot.RichTextItalic{
							Text: gotgbot.RichTextString(I.GetLocalisedString("events.fsx_disclaimer", langCode)),
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
	stopActionLoop()
	return err
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
	if len(ctx.Args()) == 1 {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
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
		_, err = b.SendRichMessageWithContext(ctxTimeout, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		log.Printf("Sent ongoing events to user %d", ctx.EffectiveUser.Id)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("events.request_timeout", langCode), nil)
			log.Printf("Request timed out while sending ongoing events to user %d", ctx.EffectiveUser.Id)
		}
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
