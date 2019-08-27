package mailgun

import (
	"fmt"
	htpl "html/template"
	"io"
	"path/filepath"
	ttpl "text/template"
	"time"

	"github.com/rojakcoder/mailgun-caddy/maillog"
)

const defaultEndpoint = "/mailgun"

type renderer interface {
	// Execute runs the template renderer and parses data (interface{}) into
	// the parsed unterlying template. The output will be written
	// to the io.Writer.
	Execute(io.Writer, interface{}) error
}

type config struct {
	// Mailgun parameters
	// domain is the domain registered with Mailgun (https://app.mailgun.com/app/domains)
	domain string
	// privatekey is the API key (https://app.mailgun.com/app/account/security)
	privatekey string

	// endpoint the route name where we receive the post requests
	endpoint string

	// maillog writes each email into one file in a directory. If nil, writes
	// to /dev/null also logs errors.
	maillog maillog.Logger

	// from            sender_from@domain.email
	fromEmail string
	fromName  string // Name of the sender

	// to              recipient_to@domain.email
	to []string
	// cc              recipient_cc1@domain.email, recipient_cc2@domain.email
	cc []string
	// bcc             recipient_bcc1@domain.email, recipient_bcc2@domain.email
	bcc []string
	// subject         Email from {{.firstname}} {{.lastname}}
	subject string

	// subjectTpl parsed and loaded subject template
	subjectTpl *ttpl.Template

	//body            path/to/tpl.[txt|html]
	body       string
	bodyIsHTML bool
	// bodyTpl parsed and loaded HTML or Text template for the email body.
	bodyTpl renderer

	rateLimitInterval time.Duration
	rateLimitCapacity int64
}

func newConfig() *config {
	return &config{
		endpoint:          defaultEndpoint,
		rateLimitInterval: time.Hour * 24,
		rateLimitCapacity: 1000,
	}
}

func (c *config) loadFromEnv() error {
	c.domain = loadFromEnv(c.domain)
	c.privatekey = loadFromEnv(c.privatekey)
	return nil
}

func (c *config) loadTemplate() (err error) {
	if !fileExists(c.body) {
		return fmt.Errorf("[mailgun] File %q not found", c.body)
	}

	switch filepath.Ext(c.body) {
	case ".txt":
		c.bodyTpl, err = ttpl.ParseFiles(c.body)
	case ".html":
		c.bodyIsHTML = true
		c.bodyTpl, err = htpl.ParseFiles(c.body)
	default:
		return fmt.Errorf("[mailgun] Only .txt and .html extensions are accepted: %q", c.body)
	}
	if err != nil {
		return fmt.Errorf("[mailgun] File %q is not readable: %s", c.body, err)
	}

	c.subjectTpl, err = ttpl.New("").Parse(c.subject)
	return
}
