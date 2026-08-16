package gacha

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	//Remember to register this command after the handler is implemented
	dispatcher.AddHandler(handlers.NewCommand("gacha", gachaHandler))
}

func gachaHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	//TODO: Implement gacha command logic here
	return nil
}
