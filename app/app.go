package app

import (
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils/mail"
	"github.com/muerwre/vault-golang/utils/notify"
)

type App struct {
	Config   *Config
	DB       *db.DB
	Mailer   *mail.Mailer
	Notifier *notify.Notifier
}

func New() (app *App, err error) {
	app = &App{}

	if app.Config, err = InitConfig(); err != nil {
		return nil, err
	}

	if app.DB, err = db.New(); err != nil {
		return nil, err
	}

	app.Notifier = new(notify.Notifier).Init(*app.DB)

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
	a.Notifier.Done()
	a.Mailer.Done()
	a.DB.Close()

	return nil
}
