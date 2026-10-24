package gacha

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strconv"
	"strings"
	"time"

	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
	"github.com/libost/gotgbot/v2/ext/handlers/filters/callbackquery"

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

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("gacha", gachaHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("gacha_"), gachaCallbackHandler))
}

func gachaHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	context, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
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

	if gachaDetailed.GachaName[qindex] == "" {
		for index, name := range gachaDetailed.GachaName {
			if name != "" {
				qindex = index
				break
			}
		}
		region := reverseIndexHelper(qindex)
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("gacha.not_in_specified_region", langCode), regionCode, region), nil)
		regionCode = reverseIndexHelper(qindex)
	}

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
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_future", langCode), days, hrs, mins, secs)
	} else if endAt > now {
		countdown := endAt - now
		secs, mins, hrs, days := events.TimeCalc(countdown)
		countdownStr = fmt.Sprintf(I.GetLocalisedString("gacha.countdown_now", langCode), days, hrs, mins, secs)
	} else {
		countdownStr = I.GetLocalisedString("gacha.countdown_ended", langCode)
	}

	startTimeStr := time.Unix(publishedAt/1000, 0).Format("2006-01-02 15:04:05")
	endTimeStr := time.Unix(endAt/1000, 0).Format("2006-01-02 15:04:05")

	var paymentButtons []gotgbot.RichMessageButton
	if len(gachaDetailed.PaymentMethods) > 0 {
		i := 0
		for _, method := range gachaDetailed.PaymentMethods {
			if method.PaymentMethod == "free" {
				paymentButtons = append(paymentButtons, gotgbot.RichMessageButton{
					Text:         gotgbot.RichTextString(fmt.Sprintf(I.GetLocalisedString("gacha.payment_button_free", langCode), method.Count)),
					CallbackData: "gacha_" + ctx.Args()[1] + "_" + strconv.Itoa(i) + "_" + regionCode,
				})
				i++
				continue
			}
			paymentButtons = append(paymentButtons, gotgbot.RichMessageButton{
				Text:         gotgbot.RichTextString(fmt.Sprintf(I.GetLocalisedString("gacha.payment_button", langCode), method.CostItemQuantity, I.GetLocalisedString(fmt.Sprintf("gacha.%s", method.PaymentMethod), langCode), method.Count)),
				CallbackData: "gacha_" + ctx.Args()[1] + "_" + strconv.Itoa(i) + "_" + regionCode,
			})
			i++
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
					Media: gotgbot.InputFileByURL("https://bestdori.com/assets/" + regionCode + "/homebanner_rip/" + gachaDetailed.BannerAssetBundleName + ".png"),
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
				},
			},
			gotgbot.InputRichBlockDivider{},
			gotgbot.InputRichBlockButtons{
				Buttons: paymentButtons,
				Align:   "center",
			},
		},
	}

	_, err = b.SendRichMessageWithContext(context, ctx.EffectiveChat.Id, richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}

func gachaCallbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.api_error", langCode), nil)
		return err
	}

	img, err := generateGachaImage(results)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("gacha.image_generation_error", langCode), nil)
		return err
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
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
