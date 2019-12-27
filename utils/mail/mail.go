package mail

import (
	"fmt"
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
	Chan   chan *gomail.Message
	Config *MailerConfig
	Dialer *gomail.Dialer
	Open   bool
	Closer gomail.SendCloser
}

func (ml *Mailer) Init(c *MailerConfig) *Mailer {
	ml.Config = c
	ml.Dialer = gomail.NewDialer(c.Host, c.Port, c.User, c.Password)
	ml.Chan = make(chan *gomail.Message, 10)

	return ml
}

func (ml *Mailer) Listen() {
	fmt.Println("MAILER STARTED")

	ml.Open = false
	var err error

	for {
		select {
		case m, ok := <-ml.Chan:

			if !ok {
				return
			}

			if !ml.Open {
				if ml.Closer, err = ml.Dialer.Dial(); err != nil {
					panic(err)
				}

				ml.Open = true
			}

			if err := gomail.Send(ml.Closer, m); err != nil {
				fmt.Println(err)
			}

		case <-time.After(30 * time.Second):
			if ml.Open {
				if err := ml.Closer.Close(); err != nil {
					panic(err)
				}

				ml.Open = false
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

	m.SetHeader("From", ml.Config.From)
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
