package app

import (
	"github.com/spf13/viper"
)

type Config struct {
	Host     string
	Port     int
	TlsFiles []string
}

func InitConfig() (*Config, error) {
	config := &Config{
		Host:     viper.GetString("Host"),
		Port:     viper.GetInt("Port"),
		TlsFiles: viper.GetStringSlice("TlsFiles"),
	}

	return config, nil
}
