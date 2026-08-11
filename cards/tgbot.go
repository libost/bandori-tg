package cards

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/characters"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("query", queryHandler))
}

func queryHandler(b *gotgbot.Bot, ctx *ext.Context) error {
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
	for _, prefix := range card.Prefix {
		if prefix != "null" {
			name = prefix
			break
		}
	}
	char, err := characters.GetCharacter(fmt.Sprint(card.CharacterID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching character data.", nil)
		return nil
	}
	characterName := char.CharacterName[0] // Temporarily using the first name in the list, i18n later.
	band, err := band.GetBand(fmt.Sprint(char.BandID))
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching band data.", nil)
		return nil
	}
	bandName := band.BandName[0] // Temporarily using the first name in the list, i18n later.
	normalPath, trainingPath, err := GetCard(param)
	if err != nil {
		ctx.EffectiveMessage.Reply(b, "Error occurred while fetching card data.", nil)
		return nil
	}
	rarity := ""
	for i := 0; i < card.Rarity; i++ {
		rarity = rarity + "★"
	}
	caption := "Card ID: " + param + "\n" +
		"Card Name: " + name + "\n" +
		"Rarity: " + rarity + "\n" +
		"Character: " + characterName + "\n" +
		"Band: " + bandName + "\n" +
		normalPath + "\n" + trainingPath
	if normalPath == "" {
		ctx.EffectiveMessage.ReplyPhoto(b, gotgbot.InputFileByURL(trainingPath), &gotgbot.SendPhotoOpts{
			Caption: caption,
		})
		return nil
	}
	if trainingPath == "" {
		ctx.EffectiveMessage.ReplyPhoto(b, gotgbot.InputFileByURL(normalPath), &gotgbot.SendPhotoOpts{
			Caption: caption,
		})
		return nil
	}
	media := []gotgbot.InputMedia{
		&gotgbot.InputMediaPhoto{
			Media:   gotgbot.InputFileByURL(normalPath),
			Caption: caption,
		},
		&gotgbot.InputMediaPhoto{
			Media: gotgbot.InputFileByURL(trainingPath),
		},
	}
	ctx.EffectiveMessage.ReplyMediaGroup(b, media, nil)
	return nil
}
