package email

import (
	"bytes"
	"fmt"
	"net/mail"
	"strconv"
	"text/template"

	"gopkg.in/gomail.v2"
)

type email struct {
	To      string
	Bcc     []string
	From    string
	Subject string
	Header  string
	Content string
	Footer  string
}

// Config email agent configuration.
type Config struct {
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
	FromName    string
	Template    string
}

// Agent for mailing
type Agent struct {
	conf *Config
	tmpl *template.Template
	dial *gomail.Dialer
}

// New creates new email agent
func New(c *Config) (*Agent, error) {
	a := &Agent{}
	a.conf = c
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return a, err
	}
	d := gomail.NewDialer(c.Host, port, c.Username, c.Password)
	// d := gomail.NewDialer("smtp.gmail.com", 587, c.Username, c.Password)

	a.dial = d

	tmpl, err := template.ParseFiles(c.Template)
	if err != nil {
		return a, fmt.Errorf("parse e-mail template failed: %v", err)
	}
	a.tmpl = tmpl
	return a, nil
}

// Send sends e-mail
func (a *Agent) Send(To string, Receivers []string, From, Subject, Header, Content, Footer string) error {
	if a.tmpl == nil {
		return fmt.Errorf("missing e-mail template file")
	}

	buff := new(bytes.Buffer)
	em := email{
		To:      To,
		From:    From,
		Subject: Subject,
		Header:  Header,
		Content: Content,
		Footer:  Footer,
	}

	fmt.Println("To: ", To)
	fmt.Println("Receivers: ", Receivers)
	if From == "" {
		from := mail.Address{Name: a.conf.FromName, Address: a.conf.FromAddress}
		em.From = from.String()
	}

	if err := a.tmpl.Execute(buff, em); err != nil {
		return fmt.Errorf("execute e-mail template failed: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", em.From)
	// m.SetHeader("To", To...)
	m.SetHeader("To", To)
	// if len(Bcc) == 1 {
	// 	m.SetHeader("Bcc", Bcc[0])
	// } else {
	// 	m.SetHeader("Bcc", Bcc...)
	// }

	fmt.Println("Bcc = ", Receivers)
	m.SetHeader("Bcc", Receivers...)
	m.SetHeader("Subject", Subject)
	m.SetBody("text/plain", buff.String())

	if err := a.dial.DialAndSend(m); err != nil {
		return fmt.Errorf("sending e-mail failed: %v", err)
	}

	return nil
}
