package main

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/cards"
	"github.com/libost/bandori-tg/characters"
	"github.com/libost/bandori-tg/commands"
	"github.com/libost/bandori-tg/config"
	"github.com/libost/bandori-tg/skills"
)

func main() {
	err := cards.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize card lists:", err)
	}
	err = characters.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize character lists:", err)
	}
	err = band.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize band lists:", err)
	}
	err = skills.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize skill lists:", err)
	}
	config.InitConfig()
	token := config.AppConfig.General.Token
	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic(err)
	}
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Printf("Error occurred while processing update %v: %v", ctx.Update, err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	commands.AddHandlers(dispatcher)
	cards.AddHandlers(dispatcher)

	updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
	})

	fmt.Println("Bot is running. Press Ctrl+C to stop.")

	updater.Idle()

}
