package mail

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

// var MailRestoreSubject = "Кто-то пытается сбросить пароль!"

type MailerConfig struct {
	From     string
	Host     string
	Port     int
	User     string
	Password string
}

type Mailer struct {
	config *MailerConfig
	dialer *gomail.Dialer
	open   bool
	closer gomail.SendCloser

	Chan chan *gomail.Message
}

func (ml *Mailer) Init(c *MailerConfig) *Mailer {
	ml.config = c
	ml.dialer = gomail.NewDialer(c.Host, c.Port, c.User, c.Password)

	ml.Chan = make(chan *gomail.Message, 10)

	return ml
}

func (ml *Mailer) Listen() {
	logrus.Info("Mailer routine started")
	logrus.Infof("Smtp relay via %s:%d", ml.config.Host, ml.config.Port)

	ml.open = false
	var err error

	for {
		select {
		case m, ok := <-ml.Chan:
			if !ok {
				logrus.Warnf("Mailer channel closed")
				return
			}

			if !ml.open {
				if ml.closer, err = ml.dialer.Dial(); err != nil {
					panic(err)
				}

				ml.open = true
			}

			if err := gomail.Send(ml.closer, m); err != nil {
				logrus.Warnf("Mailer can't send mail: %s", err.Error())
			}

		case <-time.After(30 * time.Second):
			if ml.open {
				if err := ml.closer.Close(); err != nil {
					panic(err)
				}

				ml.open = false
			}
		}
	}
}

func (ml Mailer) Create(to string, subj string, text string, html string, vals *map[string]string) *gomail.Message {
	m := gomail.NewMessage()

	if vals != nil {
		for k, v := range *vals {
			text = strings.ReplaceAll(text, fmt.Sprintf("{%s}", k), v)
			html = strings.ReplaceAll(html, fmt.Sprintf("{%s}", k), v)
		}
	}

	m.SetHeader("From", ml.config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subj)
	m.SetBody("text/plain", text)

	if html != "" {
		m.AddAlternative("text/html", html)
	}

	return m
}

func (ml Mailer) Send(m *gomail.Message) {
	ml.Chan <- m
}

func (ml Mailer) Done() {
	close(ml.Chan)
}
