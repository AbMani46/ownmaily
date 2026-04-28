package mailer

import "log"

type Message struct {
	To      string
	From    string
	Subject string
	HTMLBody string
	TextBody string
	ReplyTo  string
	Headers  map[string]string
}

type Mailer interface {
	Send(msg Message) error
	Name() string
}

type LogMailer struct{}

func (l *LogMailer) Send(msg Message) error {
	log.Printf("[LogMailer] Would send to %s | subject: %s", msg.To, msg.Subject)
	return nil
}

func (l *LogMailer) Name() string { return "log" }
