package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
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
	"github.com/libost/bandori-tg/songs"
	"github.com/libost/bandori-tg/text"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/bandori-tg/version"
	"github.com/robfig/cron/v3"
	"golang.org/x/net/proxy"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	os.MkdirAll("./logs", os.ModePerm)
	os.MkdirAll("./res", os.ModePerm)
	logFile, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	logWriter := &lumberjack.Logger{
		Filename:   "./logs/app.log",
		MaxSize:    10, // megabytes
		MaxBackups: 5,
		MaxAge:     28, // days
		Compress:   true,
	}
	multiWriter := io.MultiWriter(os.Stdout, logWriter)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
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

	err = utils.InitLists()
	if err != nil {
		log.Fatalf("Failed to initialize lists: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(loc), cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DefaultLogger)))

	config.InitConfig()
	token := config.AppConfig.General.Token
	httpClient := httpClientWithProxy(config.AppConfig)
	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		BotClient: &gotgbot.BaseBotClient{
			Client: *httpClient,
		},
	})
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
	cards.AddHandlers(dispatcher)
	songs.AddHandlers(dispatcher)
	text.AddHandlers(dispatcher)

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
	_, err := c.AddFunc("30 * * * *", utils.CronInit) // 每个小时的第30分钟执行一次
	if err != nil {
		log.Printf("Error adding cron job: %v", err)
	}
	c.Start()
	select {} // Keep the main goroutine running
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
	if err != nil {
		log.Printf("invalid webhook url: %q: %v", rawURL, err)
		panic(err)
	}
	if webhookURL.Scheme == "" || webhookURL.Host == "" {
		err = fmt.Errorf("invalid webhook url (missing scheme or host): %q", rawURL)
		log.Print(err)
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

func httpClientWithProxy(cfg *C.Config) *http.Client {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	if cfg.Proxy.Enabled {
		switch cfg.Proxy.Type {
		case "socks5":
			dialer, _ := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port), nil, proxy.Direct)
			httpClient = &http.Client{
				Timeout: time.Second * 10,
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
			}
			log.Printf("using SOCKS5 proxy at %s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
		case "http":
			proxyUrl, _ := url.Parse(fmt.Sprintf("http://%s:%d", cfg.Proxy.Host, cfg.Proxy.Port))
			httpClient = &http.Client{
				Timeout: time.Second * 10,
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
			}
			log.Printf("using HTTP proxy at %s:%d", cfg.Proxy.Host, cfg.Proxy.Port)
		default:
			log.Printf("unsupported proxy type: %s", cfg.Proxy.Type)
			panic("unsupported proxy type")
		}
	}
	return httpClient
}
