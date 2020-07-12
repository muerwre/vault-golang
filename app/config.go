package app

import (
	"github.com/spf13/viper"
	"path/filepath"
)

type Config struct {
	Debug           bool
	Host            string
	Port            int
	TlsFiles        []string
	SmtpHost        string
	SmtpPort        int
	SmtpUser        string
	SmtpPassword    string
	SmtpFrom        string
	Protocol        string
	ResetUrl        string
	PublicHost      string
	UploadPath      string
	UploadMaxSizeMb int
	GoogleApiKey    string
}

func InitConfig() (*Config, error) {
	config := &Config{
		Debug:           viper.GetBool("Debug"),
		Host:            viper.GetString("Host"),
		Port:            viper.GetInt("Port"),
		TlsFiles:        viper.GetStringSlice("TlsFiles"),
		SmtpHost:        viper.GetString("Smtp.Host"),
		SmtpPort:        viper.GetInt("Smtp.Port"),
		SmtpUser:        viper.GetString("Smtp.User"),
		SmtpPassword:    viper.GetString("Smtp.Password"),
		SmtpFrom:        viper.GetString("Smtp.From"),
		ResetUrl:        viper.GetString("Frontend.ResetUrl"),
		PublicHost:      viper.GetString("Frontend.PublicHost"),
		Protocol:        "http",
		UploadPath:      filepath.Clean(viper.GetString("Uploads.Path")),
		UploadMaxSizeMb: viper.GetInt("Uploads.MaxSizeMb") * 1024 * 1024,
		GoogleApiKey:    viper.GetString("Google.ApiKey"),
	}

	if len(config.TlsFiles) == 2 {
		config.Protocol = "https"
	}

	return config, nil
}
