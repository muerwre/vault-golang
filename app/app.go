package app

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils/mail"
)

type App struct {
	Config *Config
	DB     *db.DB
	Mailer *mail.Mailer
}

func New() (app *App, err error) {
	app = &App{}
	app.Config, err = InitConfig()

	if err != nil {
		return nil, err
	}

	app.DB, err = db.New()

	if err != nil {
		return nil, err
	}

	if app.Config.SmtpHost != "" {
		app.Mailer = new(mail.Mailer).Init(&mail.MailerConfig{
			Host:     app.Config.SmtpHost,
			Port:     app.Config.SmtpPort,
			User:     app.Config.SmtpUser,
			Password: app.Config.SmtpPassword,
			From:     app.Config.SmtpFrom,
		})
	}

	return app, err
}

func (a *App) Close() error {
	return a.DB.Close()
}
