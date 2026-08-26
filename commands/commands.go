package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/libost/bandori-tg/config"
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
	dispatcher.AddHandler(handlers.NewCommand("setadmin", setAdminHandler))
	dispatcher.AddHandler(handlers.NewCommand("setcommands", setCommandsHandler))
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
	if ctx.EffectiveChat.Type != "private" {
		displayText := I.GetLocalisedString("commands.lang_desc_group", langCode)
		ctx.EffectiveMessage.Reply(b, displayText, nil)
		return nil
	}
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

func setAdminHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	if len(ctx.Args()) == 1 {
		displayText := I.GetLocalisedString("commands.setadmin_no_token", langCode)
		ctx.EffectiveMessage.Reply(b, displayText, nil)
		return nil
	} else {
		token := ctx.Args()[1]
		if token != config.AppConfig.General.AdminToken {
			langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
			displayText := I.GetLocalisedString("commands.setadmin_invalid_token", langCode)
			ctx.EffectiveMessage.Reply(b, displayText, nil)
			return nil
		}
		_, err := DB.Init("setadmin", ctx.EffectiveUser.Id, nil)
		if err != nil {
			ctx.EditedMessage.Reply(b, I.GetLocalisedString("commands.setadmin_error", langCode), nil)
			return err
		}
		displayText := I.GetLocalisedString("commands.setadmin_success", langCode)
		ctx.EffectiveMessage.Reply(b, displayText, nil)
		return nil
	}
}

func setCommandsHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	data, err := DB.Init("get_user_group", ctx.EffectiveUser.Id, nil)
	if err != nil {
		return err
	}
	userGroup, ok := data["user_group"].(string)
	if !ok {
		ctx.EffectiveMessage.Reply(b, "Error: Invalid user group type", nil)
		return fmt.Errorf("Invalid user_group type")
	}
	if userGroup != "admin" {
		ctx.EffectiveMessage.Reply(b, "Error: You are not an admin", nil)
		return fmt.Errorf("User is not an admin")
	}
	// Do something with the admin privileges
	supportedLanguages, _, err := I.GetAllSupportedLanguages()
	if err != nil {
		return err
	}
	for idx, lang := range supportedLanguages {
		keyIndex := idx + 1
		name := lang[fmt.Sprintf("name_%d", keyIndex)]
		code := lang[fmt.Sprintf("code_%d", keyIndex)]
		codeAlt := lang[fmt.Sprintf("code_alt_%d", keyIndex)]
		if name == "" || code == "" {
			continue
		}
		languageCode, ok := normalizeTelegramLanguageCode(code, codeAlt)
		if !ok {
			log.Printf("%s", fmt.Sprintf("Skipping private commands for language %s due to invalid Telegram language code (code=%s, code_alt=%s)", name, code, codeAlt))
			continue
		}
		privateCommands := []gotgbot.BotCommand{
			{Command: "start", Description: I.GetLocalisedString("commands.setcommands_desc_list[0]", code)},
			{Command: "help", Description: I.GetLocalisedString("commands.setcommands_desc_list[1]", code)},
			{Command: "cards", Description: I.GetLocalisedString("commands.setcommands_desc_list[2]", code)},
			{Command: "events", Description: I.GetLocalisedString("commands.setcommands_desc_list[3]", code)},
			{Command: "cutoff", Description: I.GetLocalisedString("commands.setcommands_desc_list[7]", code)},
			{Command: "song", Description: I.GetLocalisedString("commands.setcommands_desc_list[8]", code)},
			{Command: "search", Description: I.GetLocalisedString("commands.setcommands_desc_list[9]", code)},
			{Command: "lang", Description: I.GetLocalisedString("commands.setcommands_desc_list[4]", code)},
			{Command: "about", Description: I.GetLocalisedString("commands.setcommands_desc_list[5]", code)},
		}
		if languageCode == "en" {
			languageCode = "" // Default language commands should be set with empty language code in Telegram
		}
		privateOpts := gotgbot.SetMyCommandsOpts{
			Scope:        gotgbot.BotCommandScopeAllPrivateChats{},
			LanguageCode: languageCode,
		}
		_, err = b.SetMyCommands(privateCommands, &privateOpts)
		if err != nil {
			return err
		}
		groupCommands := []gotgbot.BotCommand{
			{Command: "start", Description: I.GetLocalisedString("commands.setcommands_desc_list[6]", code)},
			{Command: "cards", Description: I.GetLocalisedString("commands.setcommands_desc_list[2]", code)},
			{Command: "events", Description: I.GetLocalisedString("commands.setcommands_desc_list[3]", code)},
			{Command: "cutoff", Description: I.GetLocalisedString("commands.setcommands_desc_list[7]", code)},
			{Command: "song", Description: I.GetLocalisedString("commands.setcommands_desc_list[8]", code)},
			{Command: "search", Description: I.GetLocalisedString("commands.setcommands_desc_list[9]", code)},
			{Command: "help", Description: I.GetLocalisedString("commands.setcommands_desc_list[1]", code)},
			{Command: "about", Description: I.GetLocalisedString("commands.setcommands_desc_list[5]", code)},
		}
		groupOpts := gotgbot.SetMyCommandsOpts{
			Scope:        gotgbot.BotCommandScopeAllGroupChats{},
			LanguageCode: languageCode,
		}
		_, err = b.SetMyCommands(groupCommands, &groupOpts)
		if err != nil {
			return err
		}
	}
	return nil
}
func isTelegramLanguageCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	for _, ch := range code {
		if ch < 'a' || ch > 'z' {
			return false
		}
	}
	return true
}

func normalizeTelegramLanguageCode(primaryCode, altCode string) (string, bool) {
	primaryCode = strings.ToLower(strings.TrimSpace(primaryCode))
	altCode = strings.ToLower(strings.TrimSpace(altCode))

	if isTelegramLanguageCode(altCode) {
		return altCode, true
	}
	if isTelegramLanguageCode(primaryCode) {
		return primaryCode, true
	}
	return "", false
}
