package commands

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("start", startHandler))
	dispatcher.AddHandler(handlers.NewCommand("help", helpHandler))
}

func startHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	ctx.EffectiveMessage.Reply(b, "Welcome to the Bandori Bot! Use /query <card_id> to get card information.", nil)
	return nil
}

func helpHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	ctx.EffectiveMessage.Reply(b, "Available commands:\n"+
		"/start - Start the bot and get a welcome message.\n"+
		"/help - Show this help message.\n"+
		"/query <card_id> - Query information about a specific card by its ID.", nil)
	return nil
}
