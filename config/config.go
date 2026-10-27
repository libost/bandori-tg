package config

import (
	"log"
	"os"

	"github.com/creasty/defaults"
	"github.com/goccy/go-yaml"
	C "github.com/libost/bandori-tg/constants"
)

var AppConfig *C.Config

func loadConfig(configPath string) (*C.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	cf := &C.Config{}
	if err := defaults.Set(cf); err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, cf)
	if err != nil {
		return nil, err
	}
	return cf, nil
}

func checkConfig(cf *C.Config) error {
	if cf.General.Token == "" {
		return C.ErrNoBotToken
	}
	if cf.General.AdminToken == "" {
		return C.ErrNoAdminToken
	}
	if cf.Webhook.Enabled {
		if cf.Webhook.Cert.SelfSigned {
			if cf.Webhook.Cert.CertPath == "" || cf.Webhook.Cert.KeyPath == "" {
				return C.ErrNoCertPath
			}
		}
		if cf.Webhook.URL == "" {
			return C.ErrNoWebhookURL
		}
		if cf.Webhook.Port == 0 {
			return C.ErrNoWebhookPort
		}
		if cf.Webhook.Secret == "" {
			log.Printf("Warning: Webhook secret is empty. It's recommended to set a secret token for security.")
		}
	}
	return nil
}

func InitConfig() {
	var configPath = C.ConfigPath
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		panic(err)
	}
	cf, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}
	err = checkConfig(cf)
	if err != nil {
		panic(err)
	}
	AppConfig = cf
}
