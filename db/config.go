package db

import "github.com/spf13/viper"

import "errors"

type Config struct {
	// Debug enables query logging
	Debug bool

	// URI for database in format user:password@tcp(host)/database
	URI string
}

func InitConfig() (*Config, error) {
	config := &Config{
		URI:   viper.GetString("DB.URI"),
		Debug: viper.GetBool("DB.Debug"),
	}

	if config.URI == "" {
		return nil, errors.New("Please, specify DB uri at config")
	}

	return config, nil
}
