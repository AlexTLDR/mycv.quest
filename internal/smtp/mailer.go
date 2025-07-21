package smtp

import (
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/wneessen/go-mail"
)

const defaultTimeout = 10 * time.Second

type Mailer struct {
	client *mail.Client
	from   string
}

func NewMailer(host string, port int, username, password, from string) (*Mailer, error) {
	client, err := mail.NewClient(host, mail.WithTimeout(defaultTimeout), mail.WithSMTPAuth(mail.SMTPAuthLogin), mail.WithPort(port), mail.WithUsername(username), mail.WithPassword(password))
	if err != nil {
		return nil, err
	}

	mailer := &Mailer{
		client: client,
		from:   from,
	}

	return mailer, nil
}

func (m *Mailer) SendWithComponents(recipient string, subject templ.Component, plainBody templ.Component, htmlBody templ.Component) error {
	msg := mail.NewMsg()

	err := msg.To(recipient)
	if err != nil {
		return err
	}

	err = msg.From(m.from)
	if err != nil {
		return err
	}

	// Render subject
	subjectBuf := new(bytes.Buffer)
	err = subject.Render(context.Background(), subjectBuf)
	if err != nil {
		return err
	}
	msg.Subject(strings.TrimSpace(subjectBuf.String()))

	// Render plain body
	plainBodyBuf := new(bytes.Buffer)
	err = plainBody.Render(context.Background(), plainBodyBuf)
	if err != nil {
		return err
	}
	msg.SetBodyString(mail.TypeTextPlain, plainBodyBuf.String())

	// Render HTML body if provided
	if htmlBody != nil {
		htmlBodyBuf := new(bytes.Buffer)
		err = htmlBody.Render(context.Background(), htmlBodyBuf)
		if err != nil {
			return err
		}
		msg.AddAlternativeString(mail.TypeTextHTML, htmlBodyBuf.String())
	}

	// Send with retry logic (same as original mailer)
	for i := 1; i <= 3; i++ {
		err = m.client.DialAndSend(msg)

		if nil == err {
			return nil
		}

		if i != 3 {
			time.Sleep(2 * time.Second)
		}
	}

	return err
}
