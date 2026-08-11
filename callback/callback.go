package callback

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setdlang_"), setDisplayLanguageHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setqlang_"), setQueryLanguageHandler))
}

func setDisplayLanguageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	callbackData := ctx.CallbackQuery.Data
	langCode := strings.TrimPrefix(callbackData, "setdlang_")
	_, err := ctx.CallbackQuery.Answer(b, nil)
	if err != nil {
		return err
	}
	_, err = DB.Init("set_display_language", ctx.CallbackQuery.From.Id, map[string]any{"lang_code": langCode})
	if err != nil {
		return err
	}
	_, _, err = b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.CallbackQuery.Message.GetMessageId(),
		Text:      I.GetLocalisedString("callback.setlang_success", langCode),
	})
	return nil
}

func setQueryLanguageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	callbackData := ctx.CallbackQuery.Data
	langCode := strings.TrimPrefix(callbackData, "setqlang_")
	_, err := ctx.CallbackQuery.Answer(b, nil)
	if err != nil {
		return err
	}
	_, err = DB.Init("set_query_language", ctx.CallbackQuery.From.Id, map[string]any{"lang_code": langCode})
	if err != nil {
		return err
	}
	lang, err := DB.Init("lang", ctx.CallbackQuery.From.Id, nil)
	if err != nil {
		return err
	}
	displayLangCode, ok := lang["display_language"].(string)
	if !ok {
		return fmt.Errorf("display_language has invalid type")
	}
	_, _, err = b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.CallbackQuery.Message.GetMessageId(),
		Text:      I.GetLocalisedString("callback.setlang_success", displayLangCode),
	})
	return nil
}
