package text

import (
	"log"
	"slices"
	"strings"

	"github.com/libost/bandori-tg/cards"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
	"github.com/libost/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/libost/gotgbot/v2/ext/handlers/filters/message"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewMessage(message.Text, textHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.All, callbackHandler))
}

func textHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	text := ctx.EffectiveMessage.Text
	switch true {
	case strings.HasPrefix(text, "查卡 "):
		return cards.QueryHandler(b, ctx)
	case strings.HasPrefix(text, "查活动 "):
		splitText := strings.Split(text, " ")
		if len(splitText) < 2 {
			return nil
		}
		langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
		qlangCode := I.QueryLangCodePrefer(ctx.EffectiveUser.Id, "jp")
		if len(splitText) >= 3 && slices.Contains(C.AcceptedRegions, splitText[2]) {
			qlangCode = splitText[2]
		}
		return events.SendDetailedEvent(b, ctx, splitText[1], langCode, qlangCode)
	case strings.HasPrefix(text, "查档线 "):
		splitText := strings.Split(text, " ")
		if len(splitText) < 2 {
			return nil
		}
		return events.FsxCommand(b, ctx)
	default:
		return nil
	}
}

func callbackHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	callbackData := ctx.CallbackQuery.Data
	parts := strings.Split(callbackData, "_")

	_, err := ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text:      "The Callback query '" + parts[0] + "' is not implemented yet...\nHow did you get here? Please report this to the bot developer.\n\nYou can find us using command /about",
		ShowAlert: true,
	})
	log.Printf("Received unexpected callback query: %s", callbackData)
	return err
}
