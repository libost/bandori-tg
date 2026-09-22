package text

import (
	"slices"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/libost/bandori-tg/cards"
	C "github.com/libost/bandori-tg/constants"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewMessage(message.Text, textHandler))
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
