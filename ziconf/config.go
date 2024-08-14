package ziconf

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config interface {
	GetNewRelic() NewRelicConfig
	GetService() ServiceConfig
	GetLog() LogConfig
	GetHttpPort() string
}

type NewRelicConfig struct {
	LicenseKey string `json:"license_key"`
}

type ServiceConfig struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type LogConfig struct {
	Level string `json:"level"`
}

func ReadConfig[T Config]() *T {
	var cfg T
	f := func() error {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		return viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
			dc.TagName = "json"
		})
	}

	if err := f(); err != nil {
		panic(err)
	}

	return &cfg
}
