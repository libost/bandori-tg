package callback

import (
	"fmt"
	"strconv"
	"strings"

	DB "github.com/libost/bandori-tg/database"
	I "github.com/libost/bandori-tg/i18n"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
	"github.com/libost/gotgbot/v2/ext/handlers/filters/callbackquery"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setdlang_"), setDisplayLanguageHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setqlang_"), setQueryLanguageHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setdlangid_"), setDisplayLanguageIdHandler))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("setqlangid_"), setQueryLanguageIdHandler))
}

func setDisplayLanguageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	currentlangCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	callbackData := ctx.CallbackQuery.Data
	parts := strings.Split(callbackData, "_")
	if parts[0] != "setdlang" {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", callbackData)
	}
	if len(parts) < 3 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("missing user ID in callback data: %s", callbackData)
	}
	targetUserId, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("error parsing user ID: %s", err.Error())
	}
	if targetUserId != ctx.EffectiveUser.Id {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.user_mismatch", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return nil
	}
	_, err = ctx.CallbackQuery.Answer(b, nil)
	if err != nil {
		return err
	}
	langCode := parts[1]
	_, err = DB.Init("set_display_language", ctx.CallbackQuery.From.Id, map[string]any{"lang_code": langCode})
	if err != nil {
		return err
	}
	_, _, err = b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:    ctx.EffectiveChat.Id,
		MessageId: ctx.CallbackQuery.Message.GetMessageId(),
		Text:      I.GetLocalisedString("callback.setlang_success", langCode),
	})
	return err
}

func setQueryLanguageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	currentlangCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	callbackData := ctx.CallbackQuery.Data
	parts := strings.Split(callbackData, "_")
	if parts[0] != "setqlang" {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", callbackData)
	}
	if len(parts) < 3 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("missing user ID in callback data: %s", callbackData)
	}
	targetUserId, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("error parsing user ID: %s", err.Error())
	}
	if targetUserId != ctx.EffectiveUser.Id {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.user_mismatch", currentlangCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return nil
	}
	_, errAnswer := ctx.CallbackQuery.Answer(b, nil)
	if errAnswer != nil {
		return errAnswer
	}
	langCode := parts[1]
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
	return err
}

func setDisplayLanguageIdHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	langCode := I.LangCodePrefer(ctx.EffectiveUser.Id, ctx.EffectiveUser.LanguageCode)
	data := ctx.CallbackQuery.Data
	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", data)
	}
	if parts[0] != "setdlangid" {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", data)
	}
	targetUserId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("error parsing user ID: %s", err.Error())
	}
	if targetUserId != ctx.EffectiveUser.Id {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.user_mismatch", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return nil
	}
	ctx.CallbackQuery.Answer(b, nil)
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
				CallbackData: fmt.Sprintf("setdlang_%s_%d", code, ctx.EffectiveUser.Id),
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
	data := ctx.CallbackQuery.Data
	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", data)
	}
	if parts[0] != "setqlangid" {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("invalid callback data: %s", data)
	}
	targetUserId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.invalid_callback_data", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return fmt.Errorf("error parsing user ID: %s", err.Error())
	}
	if targetUserId != ctx.EffectiveUser.Id {
		ctx.CallbackQuery.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
			Text:      I.GetLocalisedString("callback.user_mismatch", langCode),
			ShowAlert: true,
			CacheTime: 60,
		})
		return nil
	}
	ctx.CallbackQuery.Answer(b, nil)
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

	_, _, err = b.EditMessageText(&gotgbot.EditMessageTextOpts{
		ChatId:      ctx.EffectiveChat.Id,
		MessageId:   ctx.CallbackQuery.Message.GetMessageId(),
		Text:        displayText,
		ReplyMarkup: *inlineKeyboard,
	})
	return err
}
