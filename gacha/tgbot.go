package gacha

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
	"github.com/libost/gotgbot/v2/ext/handlers/filters/callbackquery"

	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/recent"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
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

func reverseIndexHelper(index int) string {
	switch index {
	case 0:
		return "jp"
	case 1:
		return "en"
	case 2:
		return "tw"
	case 3:
		return "cn"
	case 4:
		return "kr"
	default:
		return "jp"
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

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("gacha", gachaHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("gacha_"), gachaOngoingCallbackHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("gachaPay_"), gachaCallbackHandler))
}

func gachaHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	context, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	regionCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	qindex := indexHelper(regionCode)
	if len(ctx.Args()) == 1 {
		/*
			gachaMap := utils.ReadGacha()

			gachaIDs := make([]string, 0, len(gachaMap))
			for gachaID := range gachaMap {
				gachaIDs = append(gachaIDs, gachaID)
			}
			for range gachaIDs {
				randomGachaIndex, err := randomNumGen(0, len(gachaMap)-1)
				if err != nil {
					ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
					return err
				}
				gachaId = gachaIDs[randomGachaIndex]
				// Make sure the randomly selected gacha is available in the specified region
				if gachaMap[gachaId].PublishedAt[qindex] != "" && gachaMap[gachaId].ClosedAt[qindex] != "" {
					break
				}
			}

			_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.random_result", langCode), nil)
			if err != nil {
				return err
			}*/

		ongoingGachas, err := recent.GetOngoingGachas()
		if err != nil {
			log.Printf("Error fetching ongoing gachas: %v", err)
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
			return err
		}
		if len(ongoingGachas) == 0 {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.no_ongoing_gachas", langCode), nil)
			return nil
		}
		richMessage := gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(I.GetLocalisedString("gacha.ongoing_gachas", langCode)),
					Size: 3,
				},
				gotgbot.InputRichBlockDivider{},
			},
		}
		for i, gacha := range ongoingGachas {
			if len(gacha.Gacha) == 0 {
				continue
			}
			richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(I.GetLocalisedString(fmt.Sprintf("gacha.ongoing_%s", gacha.Region), langCode)),
				Size: 5,
			})
			richBlockDetails := gotgbot.InputRichBlockDetails{
				Blocks: []gotgbot.InputRichBlock{},
			}
			for _, gachaID := range gacha.Gacha {
				gacha, err := GetGacha(gachaID)
				if err != nil {
					log.Printf("Error fetching gacha details for ID %s: %v", gachaID, err)
					ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
					return err
				}
				richBlockDetails.Blocks = append(richBlockDetails.Blocks, gotgbot.InputRichBlockButtons{
					Buttons: []gotgbot.RichMessageButton{
						{
							Text:         gotgbot.RichTextString(gacha.GachaName[i]),
							Style:        "primary",
							CallbackData: "gacha_" + gachaID + fmt.Sprintf("_%d", i),
						},
					},
					Align: "center",
				})
			}
			richMessage.Blocks = append(richMessage.Blocks, richBlockDetails)
			if i == len(ongoingGachas)-1 {
				break
			}
			richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockDivider{})
		}
		_, err = b.SendRichMessageWithContext(context, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
		return err
	}
	gachaId := ctx.Args()[1]

	gacha, ok := GetGacha(gachaId)
	if ok != nil {
		log.Printf("Invalid gacha ID: %s", gachaId)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.invalid_argument", langCode), nil)
		return err
	}

	if gacha.GachaName[qindex] == "" {
		for index, name := range gacha.GachaName {
			if name != "" {
				qindex = index
				break
			}
		}
		region := reverseIndexHelper(qindex)
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("gacha.not_in_specified_region", langCode), regionCode, region), nil)
		regionCode = reverseIndexHelper(qindex)
	}

	err := sendGachaDetails(b, ctx, context, gachaId, regionCode, langCode)
	return err
}

func gachaCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
	callbackData := ctx.CallbackQuery.Data
	parts := strings.Split(callbackData, "_")
	if len(parts) < 4 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("gacha.invalid_callback_data", langCode),
			ShowAlert: true,
		})
		return fmt.Errorf("invalid callback data")
	}
	ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      I.GetLocalisedString("gacha.processing", langCode),
		ShowAlert: false,
	})
	gachaId := parts[1]
	gachaDetailed, err := GetGachaDetails(gachaId)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	regionCode := parts[3]
	paymentIndex, err := strconv.Atoi(parts[2])
	if err != nil || paymentIndex < 0 || paymentIndex >= len(gachaDetailed.PaymentMethods) {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.invalid_payment_index", langCode), nil)
		return fmt.Errorf("invalid payment index")
	}
	results, err := gachaHelper(gachaDetailed, regionCode, gachaDetailed.PaymentMethods[paymentIndex])
	if err != nil {
		log.Printf("Error generating gacha results for ID %s: %v", gachaId, err)
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}

	img, err := generateGachaImage(results)
	if err != nil {
		log.Printf("Error generating gacha image for ID %s: %v", gachaId, err)
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.image_generation_error", langCode), nil)
		return err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Printf("Error encoding gacha image for ID %s: %v", gachaId, err)
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.image_encoding_error", langCode), nil)
		return err
	}

	_, err = b.SendPhoto(ctx.EffectiveChat.Id, gotgbot.InputFileByReader("image.png", &buf), &gotgbot.SendPhotoOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}

func gachaOngoingCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	context, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	callbackData := ctx.CallbackQuery.Data
	parts := strings.Split(callbackData, "_")
	if len(parts) != 3 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("gacha.invalid_callback_data", I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)),
			ShowAlert: true,
		})
		log.Printf("Invalid callback data: %s", callbackData)
		return fmt.Errorf("invalid callback data")
	}
	gachaId := parts[1]
	region := parts[2]
	regionInt, err := strconv.Atoi(region)
	if err != nil || regionInt < 0 || regionInt >= len(C.AcceptedRegions) {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("gacha.invalid_callback_data", langCode),
			ShowAlert: true,
		})
		return fmt.Errorf("invalid region code")
	}
	ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      I.GetLocalisedString("gacha.processing", langCode),
		ShowAlert: false,
	})
	regionCode := C.AcceptedRegions[regionInt]
	err = sendGachaDetails(b, ctx, context, gachaId, regionCode, langCode)
	return err
}

func sendGachaDetails(b *gotgbot.Bot, ctx *ext.Context, context context.Context, gachaId string, regionCode string, langCode string) error {
	_, ok := GetGacha(gachaId)
	if ok != nil {
		log.Printf("Invalid gacha ID: %s", gachaId)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.invalid_argument", langCode), nil)
		return err
	}
	gachaDetailed, err := GetGachaDetails(gachaId)
	if err != nil {
		log.Printf("Error fetching gacha details for ID %s: %v", gachaId, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}

	qindex := indexHelper(regionCode)

	now := time.Now().Unix() * 1000 // Convert to milliseconds
	var countdownStr string
	publishedAt, parseErr := strconv.ParseInt(gachaDetailed.PublishedAt[qindex], 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing published time for gacha %s: %v", gachaId, parseErr)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	endAt, parseErr := strconv.ParseInt(gachaDetailed.ClosedAt[qindex], 10, 64)
	if parseErr != nil {
		log.Printf("Error parsing end time for gacha %s: %v", gachaId, parseErr)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	if publishedAt > now {
		countdown := publishedAt - now
		secs, mins, hrs, days := events.TimeCalc(countdown)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_future", langCode), days, hrs, mins, secs)
	} else if endAt > now {
		countdown := endAt - now
		secs, mins, hrs, days := events.TimeCalc(countdown)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_now", langCode), days, hrs, mins, secs)
	} else {
		countdownStr = I.GetLocalisedString("gacha.countdown_ended", langCode)
	}
	tz := tzHelper(regionCode)
	loc, _ := time.LoadLocation(tz)
	startTimeStr := time.Unix(publishedAt/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	endTimeStr := time.Unix(endAt/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	var buttons []gotgbot.InputRichBlock
	if len(gachaDetailed.PaymentMethods) > 0 {
		for i, method := range gachaDetailed.PaymentMethods {
			if method.PaymentMethod == "free" {
				buttons = append(buttons, gotgbot.InputRichBlockButtons{
					Buttons: []gotgbot.RichMessageButton{
						{
							Text:         gotgbot.RichTextString(fmt.Sprintf(I.GetLocalisedString("gacha.payment_button_free", langCode), method.Count)),
							Style:        "primary",
							CallbackData: "gacha_" + gachaId + "_" + strconv.Itoa(i) + "_" + regionCode,
						},
					},
					Align: "center",
				})
				continue
			}
			buttons = append(buttons, gotgbot.InputRichBlockButtons{
				Buttons: []gotgbot.RichMessageButton{
					{
						Text:         gotgbot.RichTextString(fmt.Sprintf(I.GetLocalisedString("gacha.payment_button", langCode), method.CostItemQuantity, I.GetLocalisedString(fmt.Sprintf("gacha.%s", method.PaymentMethod), langCode), method.Count)),
						Style:        "primary",
						CallbackData: "gachaPay_" + gachaId + "_" + strconv.Itoa(i) + "_" + regionCode,
					},
				},
				Align: "center",
			})
		}
	}
	var pickupCards []string
	for i, cards := range gachaDetailed.Details[qindex] {
		if cards.Pickup {
			pickupCards = append(pickupCards, i)
		}
	}

	if len(pickupCards) > 5 {
		pickupCards = pickupCards[:5]
	}

	img, err := generateGachaImage(pickupCards)
	if err != nil {
		log.Printf("Error generating gacha(%s) image: %v", gachaId, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		log.Printf("Error encoding image(%s): %v", gachaId, err)
		_, err := ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}
	pickupSection := gotgbot.InputRichBlockDetails{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByReader("pickup_cards.png", &buf),
				},
			},
		},
	}
	pickupHeading := gotgbot.InputRichBlockSectionHeading{
		Text: gotgbot.RichTextString(I.GetLocalisedString("gacha.pickup_cards", langCode)),
		Size: 5,
	}
	bannerUrl := "https://bestdori.com/assets/" + regionCode + "/homebanner_rip/" + gachaDetailed.BannerAssetBundleName + ".png"
	if !isValidPNGUrl(bannerUrl) {
		foundValid := false
		regionList := C.AcceptedRegions
		for _, region := range regionList {
			tempBannerUrl := "https://bestdori.com/assets/" + region + "/homebanner_rip/" + gachaDetailed.BannerAssetBundleName + ".png"
			if isValidPNGUrl(tempBannerUrl) {
				bannerUrl = tempBannerUrl
				foundValid = true
				break
			}
		}
		if !foundValid {
			log.Printf("No valid banner found for gacha %s in any region. Attempting to use logo as fallback.", gachaId)
			tempBannerUrl := "https://bestdori.com/assets/" + regionCode + "/gacha/screen/" + gachaDetailed.ResourceName + "_rip/logo.png"
			if isValidPNGUrl(tempBannerUrl) {
				bannerUrl = tempBannerUrl
				foundValid = true
			} else {
				for _, region := range regionList {
					tempBannerUrl := "https://bestdori.com/assets/" + region + "/gacha/screen/" + gachaDetailed.ResourceName + "_rip/logo.png"
					if isValidPNGUrl(tempBannerUrl) {
						bannerUrl = tempBannerUrl
						foundValid = true
						break
					}
				}
			}
			if !foundValid {
				log.Printf("No valid logo found for gacha %s in any region. Skipping.", gachaId)
			}
		}
	}
	richMessage := gotgbot.InputRichMessage{
		Blocks: []gotgbot.InputRichBlock{
			gotgbot.InputRichBlockSectionHeading{
				Text: gotgbot.RichTextString(gachaDetailed.GachaName[qindex]),
				Size: 3,
			},
			gotgbot.InputRichBlockPhoto{
				Photo: gotgbot.InputMediaPhoto{
					Media: gotgbot.InputFileByURL(bannerUrl),
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
							Text:  gotgbot.RichTextString(I.GetLocalisedString(fmt.Sprintf("gacha.%s", gachaDetailed.Type), langCode)),
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
					{
						{
							Text:  gotgbot.RichTextString(I.GetLocalisedString("gacha.id", langCode)),
							Align: "left",
						},
						{
							Text:  gotgbot.RichTextString(gachaId),
							Align: "right",
						},
					},
				},
			},
			gotgbot.InputRichBlockDivider{},
		},
	}
	if len(pickupCards) > 0 {
		richMessage.Blocks = append(richMessage.Blocks, pickupHeading, pickupSection, gotgbot.InputRichBlockDivider{})
	}
	richMessage.Blocks = append(richMessage.Blocks, buttons...)

	_, err = b.SendRichMessageWithContext(context, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}

func isValidPNGUrl(url string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "image/png" {
		return false
	}

	return true
}
