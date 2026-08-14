package cards

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/characters"
	C "github.com/libost/bandori-tg/constants"
	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/skills"
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

func isWebPageContentErr(err error) bool {
	if err == nil {
		return false
	}
	if tgErr, ok := errors.AsType[*gotgbot.TelegramError](err); ok {
		if tgErr.Code == 400 && strings.Contains(tgErr.Description, "wrong type of the web page content") {
			return true
		}
	}
	return false
}

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("query", queryHandler))
	dispatcher.AddHandler(handlers.NewCommand("查卡", queryHandler)) // 经典传承
	dispatcher.AddHandler(handlers.NewMessage(message.Text, textHandler))
}

func textHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	text := ctx.EffectiveMessage.Text
	switch true {
	case strings.HasPrefix(text, "查卡 "):
		return queryHandler(b, ctx)
	default:
		return nil
	}
}

func queryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	dlangCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.no_card_id", dlangCode), nil)
		log.Printf("User %d did not provide a card ID for query command.", ctx.EffectiveUser.Id)
		return nil
	}
	param := ctx.Args()[1]
	if len(ctx.Args()) > 2 && slices.Contains(C.AcceptedRegions, ctx.Args()[2]) {
		qlangCode = ctx.Args()[2]
		if !slices.Contains(C.AcceptedRegions, qlangCode) {
			ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.invalid_region", dlangCode), nil)
			log.Printf("User %d provided an invalid region code: %s.", ctx.EffectiveUser.Id, qlangCode)
		}
	}
	// Process the card ID and query the database
	card, exists := Cards[param]
	if !exists {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.card_not_found", dlangCode), nil)
		log.Printf("User %d queried a non-existent card with ID %s.", ctx.EffectiveUser.Id, param)
		return nil
	}
	// Process the card and send the response
	var index int
	switch qlangCode {
	case "jp":
		index = 0
	case "en":
		index = 1
	case "tw":
		index = 2
	case "cn":
		index = 3
	case "kr":
		index = 4
	}
	regionCode := regionCodeFromCard(card, qlangCode)
	if regionCode != qlangCode {
		ctx.EffectiveMessage.Reply(b, fmt.Sprintf(I.GetLocalisedString("cards.region_no_data", dlangCode), qlangCode, regionCode), nil)
	}
	cardDetailed, err := GetDetailedCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching detailed card data for card ID %s: %v", param, err)
		return nil
	}
	gachaText := strings.ReplaceAll(selectLocaleString(cardDetailed.GachaText, index), "\n", "")
	cardType := I.GetLocalisedString("cards."+card.Type, dlangCode)
	name := selectLocaleString(card.Prefix, index)

	char, err := characters.GetCharacter(fmt.Sprint(card.CharacterID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching character data for card ID %s: %v", param, err)
		return nil
	}
	characterName := selectLocaleString(char.CharacterName, index)
	band, err := band.GetBand(fmt.Sprint(char.BandID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching band data for character ID %d: %v", card.CharacterID, err)
		return nil
	}
	bandName := selectLocaleString(band.BandName, index)
	normalPath, trainingPath, err := GetCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching card data for card ID %s: %v", param, err)
		return nil
	}
	rarityEmoji := fmt.Sprint(C.RarityEmoji[0])
	if trainingPath != "" {
		rarityEmoji = fmt.Sprint(C.RarityEmoji[1])
	}
	var rarityBlock gotgbot.RichTextArray
	for i := 0; i < card.Rarity; i++ {
		rarityBlock = append(rarityBlock, gotgbot.RichTextCustomEmoji{
			CustomEmojiId:   fmt.Sprint(rarityEmoji),
			AlternativeText: "😐",
		})
	}
	skill, err := skills.GetSkill(fmt.Sprint(card.SkillID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching skill data for card ID %s: %v", param, err)
		return nil
	}
	skillName := strings.ReplaceAll(selectLocaleString(skill.SimpleDescription, index), "\n", "")
	attribute := card.Attribute
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
	releasedAt := selectLocaleString(card.ReleasedAt, index)
	releasedAtInt, _ := strconv.Atoi(releasedAt)
	var tz string
	switch qlangCode {
	case "jp":
		tz = "Asia/Tokyo"
	case "en":
		tz = "America/New_York"
	case "tw":
		tz = "Asia/Taipei"
	case "cn":
		tz = "Asia/Shanghai"
	case "kr":
		tz = "Asia/Seoul"
	default:
		tz = "UTC"
	}
	loc, _ := time.LoadLocation(tz)
	releasedAtTime := time.Unix(int64(releasedAtInt)/1000, 0).In(loc).Format("2006-01-02 15:04:05 MST")
	log.Printf("User %d queried card ID %s", ctx.EffectiveUser.Id, param)
	cemojiID := C.CharaEmoji[card.CharacterID-1]
	bemojiID := C.BandEmoji[char.BandID-1]
	richTable := &gotgbot.RichBlockTable{
		Cells: [][]gotgbot.RichBlockTableCell{
			// 1st row
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.region", dlangCode)),
					Align: "left",
				},
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards."+regionCodeFromCard(card, qlangCode), dlangCode)),
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_title", dlangCode)),
					Align: "left",
				},
				{
					Text:  gotgbot.RichTextString(name),
					Align: "right",
				},
			},
			// 2nd row
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_type", dlangCode)),
					Align: "left",
				},
				{
					Text:  gotgbot.RichTextString(cardType),
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_chara", dlangCode)),
					Align: "left",
				},
				{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(characterName + " "),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(cemojiID),
							AlternativeText: "😐",
						},
					},
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_band", dlangCode)),
					Align: "left",
				},
				{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(bandName + " "),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(bemojiID),
							AlternativeText: "😐",
						},
					},
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_attribute", dlangCode)),
					Align: "left",
				},
				{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(attribute + " "),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(C.AttributeEmoji[aindex]),
							AlternativeText: "😐",
						},
					},
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_rarity", dlangCode)),
					Align: "left",
				},
				{
					Text:  rarityBlock,
					Align: "right",
				},
			},
			{
				{
					Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_skill", dlangCode)),
					Align: "left",
				},
				{
					Text:  gotgbot.RichTextString(skillName),
					Align: "right",
				},
			},
		},
	}
	if gachaText != "" {
		richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
			{
				Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_gacha_text", dlangCode)),
				Align: "left",
			},
			{
				Text:  gotgbot.RichTextString(gachaText),
				Align: "right",
			},
		})
	}
	richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_released_at", dlangCode)),
			Align: "left",
		},
		{
			Text: gotgbot.RichTextDateTime{
				Text:           gotgbot.RichTextString(releasedAtTime),
				UnixTime:       int64(releasedAtInt) / 1000,
				DateTimeFormat: "DwT",
			},
			Align: "right",
		},
	})
	richTable.Cells = append(richTable.Cells, []gotgbot.RichBlockTableCell{
		{
			Text:  gotgbot.RichTextString(I.GetLocalisedString("cards.card_info_id", dlangCode)),
			Align: "left",
		},
		{
			Text:  gotgbot.RichTextString(param),
			Align: "right",
		},
	})
	var richMessage *gotgbot.InputRichMessage
	if normalPath == "" {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.Heading", dlangCode)),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(trainingPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
					},
				},
			},
		}
	} else if trainingPath == "" {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.Heading", dlangCode)),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(normalPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
					},
				},
			},
		}
	} else {
		richMessage = &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.Heading", dlangCode)),
					Size: 3,
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(normalPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.NormalCard", dlangCode)),
					},
				},
				gotgbot.InputRichBlockPhoto{
					Photo: gotgbot.InputMediaPhoto{
						Media: gotgbot.InputFileByURL(trainingPath),
					},
					Caption: &gotgbot.RichBlockCaption{
						Text: gotgbot.RichTextString(I.GetLocalisedString("cards.TrainingCard", dlangCode)),
					},
				},
			},
		}
	}
	richMessage.Blocks = append(richMessage.Blocks, gotgbot.InputRichBlockTable{
		Cells: richTable.Cells,
	})
	_, err = b.SendRichMessage(ctx.EffectiveUser.Id, *richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	if isWebPageContentErr(err) {
		// Handle web page content error
		richMessageErr := &gotgbot.InputRichMessage{
			Blocks: []gotgbot.InputRichBlock{
				gotgbot.InputRichBlockSectionHeading{
					Text: gotgbot.RichTextString(I.GetLocalisedString("cards.Heading", dlangCode)),
					Size: 3,
				},
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextBold{
						Text: gotgbot.RichTextItalic{
							Text: gotgbot.RichTextString(fmt.Sprintf(I.GetLocalisedString("cards.bad_card", dlangCode), param)),
						},
					},
				},
				gotgbot.InputRichBlockTable{
					Cells: richTable.Cells,
				},
			},
		}
		_, err = b.SendRichMessage(ctx.EffectiveUser.Id, *richMessageErr, &gotgbot.SendRichMessageOpts{
			ReplyParameters: &gotgbot.ReplyParameters{
				MessageId: ctx.EffectiveMessage.MessageId,
			},
		})
	}
	return err
}
