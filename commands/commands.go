package commands

import (
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	DB "github.com/libost/bandori-tg/database"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/version"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("help", helpHandler))
	dispatcher.AddHandler(handlers.NewCommand("lang", langHandler))
	dispatcher.AddHandler(handlers.NewCommand("dlang", dlangHandler))
	dispatcher.AddHandler(handlers.NewCommand("qlang", qlangHandler))
	dispatcher.AddHandler(handlers.NewCommand("about", aboutHandler))
}

func startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) > 1 {
		param := ctx.Args()[1]
		if _, ok := strings.CutPrefix(param, "events_"); ok {
			split := strings.Split(param, "_")
			eventID := split[1]
			qlangCode := split[2]
			return events.SendDetailedEvent(b, ctx, eventID, langCode, qlangCode)
		}
	}
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

func langHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	displayText := I.GetLocalisedString("commands.lang_desc", langCode)
	inlineKeyboard := &gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         I.GetLocalisedString("commands.lang_dlang", langCode),
					CallbackData: fmt.Sprintf("setdlangid_%d", ctx.EffectiveUser.Id),
					Style:        "primary",
				},
			},
			{
				{
					Text:         I.GetLocalisedString("commands.lang_qlang", langCode),
					CallbackData: fmt.Sprintf("setqlangid_%d", ctx.EffectiveUser.Id),
					Style:        "primary",
				},
			},
		},
	}
	ctx.EffectiveMessage.Reply(b, displayText, &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	},
	)
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

func aboutHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	displayText := fmt.Sprintf(I.GetLocalisedString("commands.about_desc", langCode), version.Version, version.BuildTime, version.GitCommit, version.Branch)
	ctx.EffectiveMessage.Reply(b, displayText, nil)
	return nil
}
