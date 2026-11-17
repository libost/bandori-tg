package commands

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/libost/bandori-tg/config"
	DB "github.com/libost/bandori-tg/database"
	"github.com/libost/bandori-tg/events"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/bandori-tg/version"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
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
				Text:         "International 🌐 (EN)",
				CallbackData: fmt.Sprintf("setqlang_en_%d", ctx.EffectiveUser.Id),
				Style:        "primary",
			}},
			{{
				Text:         "日本 🇯🇵 (JP)",
				CallbackData: fmt.Sprintf("setqlang_jp_%d", ctx.EffectiveUser.Id),
				Style:        "primary",
			}},
			{{
				Text:         "中国 🇨🇳 (CN)",
				CallbackData: fmt.Sprintf("setqlang_cn_%d", ctx.EffectiveUser.Id),
				Style:        "primary",
			}},
			{{
				Text:         "台灣 🇹🇼 (TW)",
				CallbackData: fmt.Sprintf("setqlang_tw_%d", ctx.EffectiveUser.Id),
				Style:        "primary",
			}},
			{{
				Text:         "한국 🇰🇷 (KR)",
				CallbackData: fmt.Sprintf("setqlang_kr_%d", ctx.EffectiveUser.Id),
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
	system := runtime.GOOS
	arch := runtime.GOARCH
	goVersion := runtime.Version()
	displayText := fmt.Sprintf(I.GetLocalisedString("commands.about_desc", langCode), version.Version, version.BuildTime, version.GitCommit, version.Branch, system, arch, goVersion)
	ctx.EffectiveMessage.Reply(b, displayText, &gotgbot.SendMessageOpts{
		ParseMode: "HTML",
	})
	return nil
}

func setAdminHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	DB.Init("create", ctx.EffectiveUser.Id, nil)
	if ctx.EffectiveChat.Type != "private" {
		log.Printf("Ignoring /setadmin command in non-private chat (chat ID: %d, chat type: %s, initiator: %d)", ctx.EffectiveChat.Id, ctx.EffectiveChat.Type, ctx.EffectiveUser.Id)
		return nil
	}
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
	if ctx.EffectiveChat.Type != "private" {
		log.Printf("Ignoring /setcommands command in non-private chat (chat ID: %d, chat type: %s, initiator: %d)", ctx.EffectiveChat.Id, ctx.EffectiveChat.Type, ctx.EffectiveUser.Id)
		return nil
	}
	data, err := DB.Init("get_user_group", ctx.EffectiveUser.Id, nil)
	if err != nil {
		return err
	}
	userGroup, ok := data["user_group"].(string)
	if !ok {
		ctx.EffectiveMessage.Reply(b, "Error: Invalid user group type", nil)
		log.Printf("Error: Invalid user group type for user %d", ctx.EffectiveUser.Id)
		return nil
	}
	if userGroup != "admin" {
		ctx.EffectiveMessage.Reply(b, "Error: You are not an admin", nil)
		log.Printf("User %d attempted to set commands but is not an admin", ctx.EffectiveUser.Id)
		return nil
	}
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
			{Command: "card", Description: I.GetLocalisedString("commands.setcommands_desc_list[2]", code)},
			{Command: "event", Description: I.GetLocalisedString("commands.setcommands_desc_list[3]", code)},
			{Command: "cutoff", Description: I.GetLocalisedString("commands.setcommands_desc_list[7]", code)},
			{Command: "comic", Description: I.GetLocalisedString("commands.setcommands_desc_list[12]", code)},
			{Command: "news", Description: I.GetLocalisedString("commands.setcommands_desc_list[13]", code)},
			{Command: "song", Description: I.GetLocalisedString("commands.setcommands_desc_list[8]", code)},
			{Command: "meta", Description: I.GetLocalisedString("commands.setcommands_desc_list[14]", code)},
			{Command: "search", Description: I.GetLocalisedString("commands.setcommands_desc_list[9]", code)},
			{Command: "char", Description: I.GetLocalisedString("commands.setcommands_desc_list[11]", code)},
			{Command: "gacha", Description: I.GetLocalisedString("commands.setcommands_desc_list[10]", code)},
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
			{Command: "help", Description: I.GetLocalisedString("commands.setcommands_desc_list[1]", code)},
			{Command: "card", Description: I.GetLocalisedString("commands.setcommands_desc_list[2]", code)},
			{Command: "event", Description: I.GetLocalisedString("commands.setcommands_desc_list[3]", code)},
			{Command: "cutoff", Description: I.GetLocalisedString("commands.setcommands_desc_list[7]", code)},
			{Command: "comic", Description: I.GetLocalisedString("commands.setcommands_desc_list[12]", code)},
			{Command: "news", Description: I.GetLocalisedString("commands.setcommands_desc_list[13]", code)},
			{Command: "song", Description: I.GetLocalisedString("commands.setcommands_desc_list[8]", code)},
			{Command: "meta", Description: I.GetLocalisedString("commands.setcommands_desc_list[14]", code)},
			{Command: "search", Description: I.GetLocalisedString("commands.setcommands_desc_list[9]", code)},
			{Command: "char", Description: I.GetLocalisedString("commands.setcommands_desc_list[11]", code)},
			{Command: "gacha", Description: I.GetLocalisedString("commands.setcommands_desc_list[10]", code)},
			{Command: "lang", Description: I.GetLocalisedString("commands.setcommands_desc_list[4]", code)},
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
	ctx.EffectiveMessage.Reply(b, "Commands set successfully for all supported languages.", nil)
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
