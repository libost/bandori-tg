package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	"github.com/libost/bandori-tg/version"
)

func main() {
	fmt.Println("started")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigs)
	args := os.Args
	if len(args) > 1 {
		switch args[1] {
		case "version":
			fmt.Printf("bandori-tg version: %s\n", version.Version)
			return
		case "help":
			fmt.Println("Usage: bandori-tg [command]")
			fmt.Println("Commands:")
			fmt.Println("  help    Show this help message")
			fmt.Println("  run     Run the bot (default)")
			return
		case "run":
			// Continue to run the bot
		default:
			fmt.Printf("Unknown command: %s\n", args[1])
			return
		}
	}
	fmt.Println("before init")
	initAll()
	fmt.Println("before config")
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
	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				if err := updater.Stop(); err != nil {
					log.Printf("Error stopping updater: %v", err)
				}
				log.Println("Bot stopped gracefully.")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, reloading configuration...")
				config.InitConfig()
			}
		}
	}()
	fmt.Println("Bot is running. Press Ctrl+C to stop.")

	updater.Idle()

}

func initAll() {
	err := cards.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize card lists:", err)
	}
	fmt.Println("before init characters")
	err = characters.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize character lists:", err)
	}
	fmt.Println("before init band")
	err = band.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize band lists:", err)
	}
	fmt.Println("before init skills")
	err = skills.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize skill lists:", err)
	}
	fmt.Println("before init events")
	err = events.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize event lists:", err)
	}
	fmt.Println("before init recent")
	err = recent.InitLists()
	if err != nil {
		log.Fatal("Failed to initialize recent lists:", err)
	}
	fmt.Println("before init database")
	_, err = DB.Init("init", 0, nil) // Initialize the database for user ID 0 (default)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}
