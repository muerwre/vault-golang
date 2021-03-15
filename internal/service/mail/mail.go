package mail

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"gopkg.in/gomail.v2"
)

type MailerConfig struct {
	From     string
	Host     string
	Port     int
	User     string
	Password string
}

type MailService struct {
	config MailerConfig
	dialer *gomail.Dialer
	log    *logrus.Logger
	open   bool
	closer gomail.SendCloser

	Chan chan *gomail.Message
}

func (ml *MailService) Init(c MailerConfig, log *logrus.Logger) *MailService {
	ml.config = c
	ml.log = log
	ml.dialer = gomail.NewDialer(c.Host, c.Port, c.User, c.Password)

	ml.Chan = make(chan *gomail.Message, 10)

	return ml
}

func (ml *MailService) Listen(ctx context.Context) {
	ml.log.Info("MailService routine started")
	ml.log.Infof("Smtp relay via %s:%d", ml.config.Host, ml.config.Port)

	ml.open = false
	var err error

	for {
		select {
		case <-ctx.Done():
			close(ml.Chan)
			ml.log.Info("MailService stopped")
			return
		case m, ok := <-ml.Chan:
			if !ok {
				ml.log.Warnf("MailService channel closed")
				return
			}

			if !ml.open {
				if ml.closer, err = ml.dialer.Dial(); err != nil {
					ml.log.Warnf(err.Error())
					continue
				}

				ml.open = true
			}

			if err := gomail.Send(ml.closer, m); err != nil {
				ml.log.Warnf("MailService can't send mail: %s", err.Error())
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

func (ml MailService) CreateMessage(to string, subj string, text string, html string, values *map[string]string) *gomail.Message {
	m := gomail.NewMessage()

	if values != nil {
		for k, v := range *values {
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

func (ml MailService) Send(m *gomail.Message) {
	ml.Chan <- m
}
