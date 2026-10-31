package characters

import (
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/libost/gotgbot/v2/ext/handlers"
)

func AddHandlers(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandler(handlers.NewCommand("character", characterCommand))
	dispatcher.AddHandler(handlers.NewCommand("char", characterCommand))
}

func characterCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	// Implementation of the character command goes here
	return nil
}
