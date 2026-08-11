package cards

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/characters"
	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/skills"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("query", queryHandler))
}

func queryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		ctx.EffectiveMessage.Reply(b, "Please provide a card ID to query.", nil)
		return nil
	}
	param := ctx.Args()[1]
	// Process the card ID and query the database
	card, exists := Cards[param]
	if !exists {
		ctx.EffectiveMessage.Reply(b, "Card not found.", nil)
		return nil
	}
	// Process the card and send the response
	var name string
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
	if card.Prefix[index] != "" && card.Prefix[index] != "null" {
		name = card.Prefix[index]
	} else {
		for _, prefix := range card.Prefix {
			if prefix != "null" {
				name = prefix
				break
			}
		}
	}

	char, err := characters.GetCharacter(fmt.Sprint(card.CharacterID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching character data.", nil)
		return nil
	}
	var characterName string
	if char.CharacterName[index] != "" && char.CharacterName[index] != "null" {
		characterName = char.CharacterName[index]
	} else {
		for _, cname := range char.CharacterName {
			if cname != "null" {
				characterName = cname
				break
			}
		}
	}
	band, err := band.GetBand(fmt.Sprint(char.BandID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching band data.", nil)
		return nil
	}
	var bandName string
	if band.BandName[index] != "" && band.BandName[index] != "null" {
		bandName = band.BandName[index]
	} else {
		for _, bname := range band.BandName {
			if bname != "null" {
				bandName = bname
				break
			}
		}
	}
	normalPath, trainingPath, err := GetCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching card data.", nil)
		return nil
	}
	rarity := ""
	for i := 0; i < card.Rarity; i++ {
		rarity = rarity + "★"
	}
	skill, err := skills.GetSkill(fmt.Sprint(card.SkillID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching skill data.", nil)
		return nil
	}
	var skillName string
	if skill.SimpleDescription[index] != "" && skill.SimpleDescription[index] != "null" {
		skillName = skill.SimpleDescription[index]
	} else {
		for _, sname := range skill.SimpleDescription {
			if sname != "null" {
				skillName = sname
				break
			}
		}
	}
	caption := "Card ID: " + param + "\n" +
		"Card Name: " + name + "\n" +
		"Rarity: " + rarity + "\n" +
		"Character: " + characterName + "\n" +
		"Band: " + bandName + "\n" +
		"Skill: " + skillName
	if normalPath == "" {
		ctx.EffectiveMessage.ReplyPhoto(b, gotgbot.InputFileByURL(trainingPath), &gotgbot.SendPhotoOpts{
			Caption:   caption,
			ParseMode: "HTML",
		})
		return nil
	}
	if trainingPath == "" {
		ctx.EffectiveMessage.ReplyPhoto(b, gotgbot.InputFileByURL(normalPath), &gotgbot.SendPhotoOpts{
			Caption:   caption,
			ParseMode: "HTML",
		})
		return nil
	}
	media := []gotgbot.InputMedia{
		&gotgbot.InputMediaPhoto{
			Media:     gotgbot.InputFileByURL(normalPath),
			Caption:   caption,
			ParseMode: "HTML",
		},
		&gotgbot.InputMediaPhoto{
			Media: gotgbot.InputFileByURL(trainingPath),
		},
	}
	ctx.EffectiveMessage.ReplyMediaGroup(b, media, nil)
	return nil
}
