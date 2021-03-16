package app

import (
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/service/mail"
	"github.com/muerwre/vault-golang/internal/service/notification/controller"
	controller2 "github.com/muerwre/vault-golang/internal/service/vk/controller"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config   *Config
	DB       *db.DB
	Mailer   *mail.MailService
	Notifier *controller.NotificationService
	Vk       *controller2.VkNotificationService
	Logger   *logrus.Logger
}

func New() (app *App, err error) {
	app = &App{
		Logger: logrus.New(),
	}

	if app.Config, err = InitConfig(); err != nil {
		return nil, err
	}

	if app.DB, err = db.New(); err != nil {
		return nil, err
	}

	app.Notifier = new(controller.NotificationService).Init(*app.DB, app.Logger, app.Config.Notifications)

	if app.Config.Mail.Host != "" {
		app.Mailer = new(mail.MailService).Init(app.Config.Mail, app.Logger)
	}

	app.Vk = controller2.New(app.Config.Notifications.Vk, *app.DB, app.Logger)

	return app, err
}

func (a *App) Close() error {
	a.DB.Close()
	return nil
}
