package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("help", helpHandler))
	dispatcher.AddHandler(handlers.NewCommand("dlang", dlangHandler))
	dispatcher.AddHandler(handlers.NewCommand("qlang", qlangHandler))
}

func startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	displayText := I.GetLocalisedString("commands.start_desc", langCode)
	ctx.EffectiveMessage.Reply(b, displayText, nil)
	return nil
}

func helpHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	displayText := I.GetLocalisedString("commands.help_desc", langCode)
	ctx.EffectiveMessage.Reply(b, displayText, nil)
	return nil
}

func dlangHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
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
	_, err = ctx.EffectiveMessage.Reply(b, displayText, &gotgbot.SendMessageOpts{
		ReplyMarkup:    inlineKeyboard,
		ProtectContent: true,
	})
	return err

}

func qlangHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
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

	_, err := ctx.EffectiveMessage.Reply(b, displayText, &gotgbot.SendMessageOpts{
		ReplyMarkup:    inlineKeyboard,
		ProtectContent: true,
	})
	return err

}
