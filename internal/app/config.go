package app

import (
	"github.com/muerwre/vault-golang/internal/service/jwt"
	"github.com/muerwre/vault-golang/internal/service/mail"
	controller2 "github.com/muerwre/vault-golang/internal/service/notification/controller"
	"github.com/muerwre/vault-golang/internal/service/vk/controller"
	"github.com/spf13/viper"
	"path/filepath"
)

type Config struct {
	Debug              bool
	ApiDebug           bool
	Port               int
	TlsFiles           []string
	Protocol           string
	ResetUrl           string
	PublicHost         string
	UploadPath         string
	UploadMaxSizeMb    int
	UploadOutputWebp   bool
	GoogleApiKey       string
	VkClientId         string
	VkClientSecret     string
	VkCallbackUrl      string
	GoogleClientId     string
	GoogleClientSecret string
	GoogleCallbackUrl  string
	Notifications      controller2.NotificationsConfig
	Mail               mail.MailerConfig
}

func InitConfig() (*Config, error) {
	config := &Config{
		Debug:              viper.GetBool("Debug"),
		ApiDebug:           viper.GetBool("Api.Debug"),
		Port:               viper.GetInt("Port"),
		TlsFiles:           viper.GetStringSlice("TlsFiles"),
		ResetUrl:           viper.GetString("Frontend.ResetUrl"),
		PublicHost:         viper.GetString("Frontend.PublicHost"),
		Protocol:           "http",
		UploadPath:         filepath.Clean(viper.GetString("Uploads.Path")),
		UploadMaxSizeMb:    viper.GetInt("Uploads.MaxSizeMb") * 1024 * 1024,
		UploadOutputWebp:   viper.GetBool("Uploads.OutputWebp"),
		GoogleApiKey:       viper.GetString("Google.ApiKey"),
		VkClientId:         viper.GetString("Vk.ClientId"),
		VkClientSecret:     viper.GetString("Vk.ClientSecret"),
		VkCallbackUrl:      viper.GetString("Vk.CallbackUrl"),
		GoogleClientId:     viper.GetString("Google.ClientId"),
		GoogleClientSecret: viper.GetString("Google.ClientSecret"),
		GoogleCallbackUrl:  viper.GetString("Google.CallbackUrl"),
		Notifications: controller2.NotificationsConfig{
			Vk: controller.VkNotificationsConfig{
				Enabled:        viper.GetBool("Notifications.Vk.Enabled"),
				ApiKey:         viper.GetString("Notifications.Vk.ApiKey"),
				GroupId:        viper.GetUint("Notifications.Vk.GroupId"),
				Delay:          viper.GetUint("Notifications.Vk.Delay"),
				CooldownMins:   viper.GetUint("Notifications.Vk.CooldownMins"),
				PurgeAfterDays: viper.GetUint("Notifications.Vk.PurgeAfterDays"),
				UrlPrefix:      viper.GetString("Notifications.Vk.UrlPrefix"),
				UploadPath:     filepath.Clean(viper.GetString("Uploads.Path")),
			},
		},
		Mail: mail.MailerConfig{
			Host:     viper.GetString("Smtp.Host"),
			Port:     viper.GetInt("Smtp.Port"),
			User:     viper.GetString("Smtp.User"),
			Password: viper.GetString("Smtp.Password"),
			From:     viper.GetString("Smtp.From"),
		},
	}

	if len(config.TlsFiles) == 2 {
		config.Protocol = "https"
	}

	jwt.InitJwtEngine(viper.GetString("Jwt.Secret"))

	return config, nil
}
