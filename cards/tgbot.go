package cards

import (
	"fmt"
	"log"
	"strings"

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
	rarity := ""
	for i := 0; i < card.Rarity; i++ {
		rarity = rarity + "★"
	}
	skill, err := skills.GetSkill(fmt.Sprint(card.SkillID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, I.GetLocalisedString("cards.error_occurred", dlangCode), nil)
		log.Printf("Error fetching skill data for card ID %s: %v", param, err)
		return nil
	}
	skillName := strings.ReplaceAll(selectLocaleString(skill.SimpleDescription, index), "\n", "")
	attribute := card.Attribute
	caption_bc := fmt.Sprintf(I.GetLocalisedString("cards.card_info_before_chara", dlangCode), name, cardType, characterName)
	caption_ac := fmt.Sprintf(I.GetLocalisedString("cards.card_info_after_chara", dlangCode), bandName)
	caption_ab := fmt.Sprintf(I.GetLocalisedString("cards.card_info_after_band", dlangCode), attribute, rarity, skillName)
	if gachaText != "" {
		caption_ab += fmt.Sprintf(I.GetLocalisedString("cards.card_info_gacha_text", dlangCode), gachaText)
	}
	caption_ab += fmt.Sprintf(I.GetLocalisedString("cards.card_info_id", dlangCode), param)
	log.Printf("User %d queried card ID %s", ctx.EffectiveUser.Id, param)
	cemojiID := C.CharaEmoji[card.CharacterID-1]
	bemojiID := C.BandEmoji[char.BandID-1]
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(caption_bc),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(cemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ac),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(bemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ab),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(caption_bc),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(cemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ac),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(bemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ab),
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
				gotgbot.InputRichBlockParagraph{
					Text: gotgbot.RichTextArray{
						gotgbot.RichTextString(caption_bc),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(cemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ac),
						gotgbot.RichTextCustomEmoji{
							CustomEmojiId:   fmt.Sprint(bemojiID),
							AlternativeText: "😐",
						},
						gotgbot.RichTextString(caption_ab),
					},
				},
			},
		}
	}
	_, err = b.SendRichMessage(ctx.EffectiveUser.Id, *richMessage, &gotgbot.SendRichMessageOpts{
		ReplyParameters: &gotgbot.ReplyParameters{
			MessageId: ctx.EffectiveMessage.MessageId,
		},
	})
	return err
}
