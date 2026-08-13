package main

import (
	"fmt"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/libost/bandori-tg/band"
	"github.com/libost/bandori-tg/callback"
	"github.com/libost/bandori-tg/cards"
	"github.com/libost/bandori-tg/characters"
	"github.com/libost/bandori-tg/commands"
	"github.com/libost/bandori-tg/config"
	DB "github.com/libost/bandori-tg/database"
	"github.com/libost/bandori-tg/events"
	"github.com/libost/bandori-tg/recent"
	"github.com/libost/bandori-tg/skills"
)

func main() {
	initAll()
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

	callback.AddHandlers(dispatcher)
	commands.AddHandlers(dispatcher)
	events.AddHandlers(dispatcher)
	cards.AddHandlers(dispatcher) // 必须放在最后，因为它包含了一个通配的消息处理器，会捕获所有未被其他处理器处理的消息。

	updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
	})

	fmt.Println("Bot is running. Press Ctrl+C to stop.")

	updater.Idle()

}

func initAll() {
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
	err = events.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize event lists:", err)
	}
	err = recent.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize recent lists:", err)
	}
	_, err = DB.Init("init", 0, nil) // Initialize the database for user ID 0 (default)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}
