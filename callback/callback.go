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
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setdlangid_"), setDisplayLanguageIdHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setqlangid_"), setQueryLanguageIdHandler))
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

func setDisplayLanguageIdHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	supportedLanguages, _, err := I.GetAllSupportedLanguages()
	if err != nil {
		return err
	}
	displayText := I.GetLocalisedString("commands.dlanguages_desc", langCode)
	inlineKeyboard := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{},
	}
	for idx, lang := range supportedLanguages {
		keyIndex := idx + 1
		name := lang[fmt.Sprintf("name_%d", keyIndex)]
		code := lang[fmt.Sprintf("code_%d", keyIndex)]
		if name == "" || code == "" {
			continue
		}
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []gotgbot.InlineKeyboardButton{
			{
				Text:         name,
				CallbackData: fmt.Sprintf("setdlang_%s", code),
				Style:        "primary",
			},
		})
	}
	_, _, err = b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.CallbackQuery.Message.GetMessageId(),
		Text:        displayText,
		ReplyMarkup: *inlineKeyboard,
	})
	return err
}

func setQueryLanguageIdHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	displayText := I.GetLocalisedString("commands.qlanguages_desc", langCode)
	inlineKeyboard := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{{
				Text:         "English",
				CallbackData: "setqlang_en",
				Style:        "primary",
			}},
			{{
				Text:         "日本語",
				CallbackData: "setqlang_jp",
				Style:        "primary",
			}},
			{{
				Text:         "中文简体",
				CallbackData: "setqlang_cn",
				Style:        "primary",
			}},
			{{
				Text:         "中文繁體",
				CallbackData: "setqlang_tw",
				Style:        "primary",
			}},
			{{
				Text:         "한국어",
				CallbackData: "setqlang_kr",
				Style:        "primary",
			}},
		},
	}

	_, _, err := b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.CallbackQuery.Message.GetMessageId(),
		Text:        displayText,
		ReplyMarkup: *inlineKeyboard,
	})
	return err
}
