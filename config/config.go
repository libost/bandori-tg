package config

import (
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
	AppConfig = cf
}
