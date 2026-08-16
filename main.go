package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/libost/bandori-tg/callback"
	"github.com/libost/bandori-tg/cards"
	"github.com/libost/bandori-tg/commands"
	"github.com/libost/bandori-tg/config"
	C "github.com/libost/bandori-tg/constants"
	DB "github.com/libost/bandori-tg/database"
	"github.com/libost/bandori-tg/events"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/bandori-tg/version"
	"github.com/robfig/cron/v3"
)

func main() {
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
	initAll()

	_ = utils.InitLists()

	loc, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(loc), cron.WithChain(cron.Recover(cron.DefaultLogger)))

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

	if config.AppConfig.Webhook.Enabled {
		startWebhookCommunication(config.AppConfig, updater, b)
	} else {
		updater.StartPolling(b, &ext.PollingOpts{
			DropPendingUpdates: true,
		})
	}

	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				if err := updater.Stop(); err != nil {
					log.Printf("Error stopping updater: %v", err)
				}
				ctx := c.Stop()
				<-ctx.Done()
				log.Println("Bot stopped gracefully.")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, reloading configuration...")
				config.InitConfig()
			}
		}
	}()
	fmt.Println("Bot is running. Press Ctrl+C to stop.")
	cronTasks(c)

	updater.Idle()

}

func initAll() {
	_, err := DB.Init("init", 0, nil) // Initialize the database for user ID 0 (default)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
}

func cronTasks(c *cron.Cron) {
	_, err := c.AddFunc("30 8 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing lists via cron: %v", err)
		}
	}) // 每天 8:30 AM 更新列表 国际服活动开始前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("30 12 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing events via cron: %v", err)
		}
	}) // 每天 12:30 PM 更新活动 国服活动开始前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("30 13 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing lists via cron: %v", err)
		}
	}) // 每天 13:30 PM 更新列表 日服活动开始前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("29 14 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing lists via cron: %v", err)
		}
	}) // 每天 14:29 PM 更新列表 国际服活动结束前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("29 19 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing events via cron: %v", err)
		}
	}) // 每天 19:29 PM 更新活动 日服活动结束前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("29 20 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing lists via cron: %v", err)
		}
	}) // 每天 20:29 PM 更新列表 台服活动结束前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	_, err = c.AddFunc("29 22 * * *", func() {
		if err := utils.InitLists(); err != nil {
			log.Printf("Error refreshing events via cron: %v", err)
		}
	}) // 每天 22:29 PM 更新活动 国服活动结束前30分钟
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	c.Start()
}

func startWebhookCommunication(cfg *C.Config, updater *ext.Updater, b *gotgbot.Bot) {
	webhookURL := mustParseWebhookURL(cfg.Webhook.URL)
	webhookOpts, setWebhookOpts := buildWebhookOptions(cfg)

	if err := updater.StartWebhook(b, webhookURL.Path, webhookOpts); err != nil {
		log.Printf("failed to start webhook: %v", err)
		panic(err)
	}

	if _, err := b.SetWebhook(cfg.Webhook.URL, setWebhookOpts); err != nil {
		log.Printf("failed to set webhook: %v", err)
		panic(err)
	}
}

func mustParseWebhookURL(rawURL string) *url.URL {
	webhookURL, err := url.Parse(rawURL)
	if err != nil || webhookURL.Scheme == "" || webhookURL.Host == "" {
		log.Printf("invalid webhook url: %q", rawURL)
		panic(err)
	}
	if webhookURL.Path == "" || webhookURL.Path == "/" {
		log.Printf("webhook url path is empty, please set a path like /webhook")
		panic("webhook url path is empty")
	}
	return webhookURL
}

func buildWebhookOptions(cfg *C.Config) (ext.WebhookOpts, *gotgbot.SetWebhookOpts) {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Webhook.Port)
	webhookOpts := ext.WebhookOpts{
		ListenAddr:  listenAddr,
		SecretToken: cfg.Webhook.Secret,
	}
	setWebhookOpts := &gotgbot.SetWebhookOpts{
		SecretToken:        webhookOpts.SecretToken,
		DropPendingUpdates: true,
	}

	if !cfg.Webhook.NginxEnabled && slices.Contains(C.TelegramAcceptPorts, cfg.Webhook.Port) {
		log.Printf("Warning: The webhook port %d is one of the ports accepted by Telegram. If you are running this bot behind a reverse proxy (like Nginx), please ensure that the proxy forwards requests to this port correctly.", cfg.Webhook.Port)
	}

	if cfg.Webhook.NginxEnabled && os.Getenv("IN_DOCKER") != "true" {
		webhookOpts.ListenAddr = fmt.Sprintf("127.0.0.1:%d", cfg.Webhook.Port)
		return webhookOpts, setWebhookOpts
	}
	return webhookOpts, setWebhookOpts
}
