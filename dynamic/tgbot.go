package dynamic

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	//Remember to register this command after the handler is implemented
	dispatcher.AddHandler(handlers.NewCommand("dynamic", dynamicHandler))
}

func dynamicHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	//TODO: Implement the dynamic command handler logic here
	return nil
}
