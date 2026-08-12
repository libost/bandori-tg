package cards

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/characters"
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
	skillName := selectLocaleString(skill.SimpleDescription, index)
	attribute := card.Attribute
	caption := fmt.Sprintf(I.GetLocalisedString("cards.card_info", dlangCode), param, name, characterName, bandName, attribute, rarity, skillName)
	log.Printf("User %d queried card ID %s", ctx.EffectiveUser.Id, param)
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
