package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/libost/bandori-tg/callback"
	"github.com/libost/bandori-tg/cards"
	"github.com/libost/bandori-tg/characters"
	"github.com/libost/bandori-tg/comic"
	"github.com/libost/bandori-tg/commands"
	"github.com/libost/bandori-tg/config"
	C "github.com/libost/bandori-tg/constants"
	DB "github.com/libost/bandori-tg/database"
	"github.com/libost/bandori-tg/events"
	"github.com/libost/bandori-tg/gacha"
	"github.com/libost/bandori-tg/songs"
	"github.com/libost/bandori-tg/text"
	"github.com/libost/bandori-tg/utils"
	"github.com/libost/bandori-tg/version"
	"github.com/libost/gotgbot/v2"
	"github.com/libost/gotgbot/v2/ext"
	"github.com/netresearch/go-cron"
	"golang.org/x/net/proxy"
	"gopkg.in/natefinch/lumberjack.v2"

	_ "time/tzdata"
)

func main() {
	os.MkdirAll("./logs", os.ModePerm)
	os.MkdirAll("./res", os.ModePerm)
	logFile, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log file: %v", err))
		os.Exit(1)
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
	slog.SetDefault(slog.New(slog.NewTextHandler(multiWriter, &slog.HandlerOptions{AddSource: true})))
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigs)
	args := os.Args
	if len(args) > 1 {
		switch args[1] {
		case "version":
			fmt.Printf("bandori-tg version: %s\n", version.Version)
			if utils.FontsDetached {
				fmt.Println("Build tags: fonts_detached")
			}
			return
		case "help":
			fmt.Println("Usage: bandori-tg [command]")
			fmt.Println("Commands:")
			fmt.Println("  help    Show this help message")
			fmt.Println("  version Show the version of the bot")
			return
		default:
			fmt.Printf("Unknown command: %s\n", args[1])
			return
		}
	}
	initDB()

	loc, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(cron.WithLocation(loc), cron.WithChain(cron.Recover(cron.DefaultLogger), cron.SkipIfStillRunning(cron.DefaultLogger)))

	config.InitConfig()
	err = utils.InitHttpClient()
	if err != nil {
		slog.Error("Failed to initialize HTTP client: ", "error", err.Error())
		os.Exit(1)
	}
	err = utils.InitLists()
	if err != nil {
		slog.Error("Failed to initialize lists: ", "error", err.Error())
		os.Exit(1)
	}
	if utils.FontsDetached {
		slog.Info("Fonts are in detached mode. Initializing font files...")
		err = utils.InitFontsFiles()
		if err != nil {
			slog.Error("Failed to initialize font files: ", "error", err.Error())
			os.Exit(1)
		}
	}
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
			slog.Error(fmt.Sprintf("Error occurred while processing update %v: %v", ctx.Update, err))
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
	gacha.AddHandlers(dispatcher)
	comic.AddHandlers(dispatcher)
	characters.AddHandlers(dispatcher)
	text.AddHandlers(dispatcher) // text handlers must be added after all other handlers to avoid conflicts

	if config.AppConfig.Webhook.Enabled {
		startWebhookCommunication(config.AppConfig, updater, b)
	} else {
		updater.StartPolling(b, &ext.PollingOpts{
			DropPendingUpdates:    true,
			EnableWebhookDeletion: true,
		})
	}

	go func() {
		for sig := range sigs {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				if err := updater.Stop(); err != nil {
					slog.Error(fmt.Sprintf("Error stopping updater: %v", err))
				}
				ctx := c.Stop()
				<-ctx.Done()
				slog.Info("Bot stopped gracefully.")
				os.Exit(0)
			case syscall.SIGHUP:
				slog.Info("Received SIGHUP, reloading configuration...")
				config.InitConfig()
			}
		}
	}()
	slog.Info("Bot is running.")
	cronTasks(c)

	updater.Idle()
}

func initDB() {
	_, err := DB.Init("init", 0, nil) // Initialize the database for user ID 0 (default)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}
}

func cronTasks(c *cron.Cron) {
	_, err := c.AddFunc("30 * * * *", utils.CronInit) // 每个小时的第30分钟执行一次
	if err != nil {
		slog.Error(fmt.Sprintf("Error adding cron job: %v", err))
	}
	c.Start()
	select {} // Keep the main goroutine running
}

func startWebhookCommunication(cfg *C.Config, updater *ext.Updater, b *gotgbot.Bot) {
	webhookURL := mustParseWebhookURL(cfg.Webhook.URL)
	webhookOpts, setWebhookOpts, pubCert := buildWebhookOptions(cfg)
	if pubCert != nil {
		defer pubCert.Close()
	}

	if err := updater.StartWebhook(b, webhookURL.Path, webhookOpts); err != nil {
		slog.Error(fmt.Sprintf("failed to start webhook: %v", err))
		panic(err)
	}

	if _, err := b.SetWebhook(cfg.Webhook.URL, setWebhookOpts); err != nil {
		slog.Error(fmt.Sprintf("failed to set webhook: %v", err))
		panic(err)
	}
}

func mustParseWebhookURL(rawURL string) *url.URL {
	webhookURL, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("invalid webhook url: %q: %v", rawURL, err)
		panic(err)
	}
	if webhookURL.Scheme == "" || webhookURL.Host == "" {
		err = fmt.Errorf("invalid webhook url (missing scheme or host): %q", rawURL)
		slog.Error(err.Error())
		panic(err)
	}
	if webhookURL.Path == "" || webhookURL.Path == "/" {
		slog.Error("webhook url path is empty, please set a path like /webhook")
		panic("webhook url path is empty")
	}
	return webhookURL
}

func buildWebhookOptions(cfg *C.Config) (ext.WebhookOpts, *gotgbot.SetWebhookOpts, *os.File) {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", cfg.Webhook.Port)
	webhookOpts := ext.WebhookOpts{
		ListenAddr:  listenAddr,
		SecretToken: cfg.Webhook.Secret,
	}
	setWebhookOpts := &gotgbot.SetWebhookOpts{
		SecretToken:        webhookOpts.SecretToken,
		DropPendingUpdates: true,
	}
	var pubCert *os.File
	if cfg.Webhook.Cert.SelfSigned {
		pubCert, err := os.Open(cfg.Webhook.Cert.CertPath)
		if err != nil {
			slog.Error("failed to open certificate file: ", "error", err.Error())
			os.Exit(1)
		}
		setWebhookOpts.Certificate = gotgbot.InputFileByReader("public.pem", pubCert)
		webhookOpts.CertFile = cfg.Webhook.Cert.CertPath
		webhookOpts.KeyFile = cfg.Webhook.Cert.KeyPath
	}

	if !cfg.Webhook.NginxEnabled && slices.Contains(C.TelegramAcceptPorts[:], cfg.Webhook.Port) {
		slog.Warn("The webhook port is one of the ports accepted by Telegram. If you are running this bot behind a reverse proxy (like Nginx), please ensure that the proxy forwards requests to this port correctly.", "port", cfg.Webhook.Port)
	}

	if cfg.Webhook.NginxEnabled && os.Getenv("IN_DOCKER") != "true" {
		webhookOpts.ListenAddr = fmt.Sprintf("127.0.0.1:%d", cfg.Webhook.Port)
		return webhookOpts, setWebhookOpts, pubCert
	}
	return webhookOpts, setWebhookOpts, pubCert
}

func httpClientWithProxy(cfg *C.Config) *http.Client {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}
	if cfg.Proxy.Enabled {
		switch cfg.Proxy.Type {
		case "socks5":
			dialer, _ := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", cfg.Proxy.Host, cfg.Proxy.Port), nil, proxy.Direct)
			httpClient = &http.Client{
				Timeout: time.Second * 30,
				Transport: &http.Transport{
					Dial: dialer.Dial,
				},
			}
			slog.Info(fmt.Sprintf("using SOCKS5 proxy at %s:%d", cfg.Proxy.Host, cfg.Proxy.Port))
		case "http":
			proxyUrl, _ := url.Parse(fmt.Sprintf("http://%s:%d", cfg.Proxy.Host, cfg.Proxy.Port))
			httpClient = &http.Client{
				Timeout: time.Second * 30,
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyUrl),
				},
			}
			slog.Info(fmt.Sprintf("using HTTP proxy at %s:%d", cfg.Proxy.Host, cfg.Proxy.Port))
		default:
			slog.Error(fmt.Sprintf("unsupported proxy type: %s", cfg.Proxy.Type))
			panic("unsupported proxy type")
		}
	}
	return httpClient
}
