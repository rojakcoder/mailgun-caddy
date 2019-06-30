package mailgun

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rojakcoder/mailgun-caddy/bufpool"
)

// Message holds the information necessary to deliver an email.
//
// It is made by piecing together information from the configuration and the request.
type Message struct {
	subject string
	from    string
	body    string
	// to is not used in the actual sending - it is for logging only
	to string
}

func newMessage(cfg *config, r *http.Request) Message {
	m := Message{}
	m.subject = makeSubject(cfg, r)
	m.body = makeBody(cfg, r)
	m.from = makeFrom(cfg, r)
	m.to = makeTo(cfg)
	return m
}

func (m *Message) copy() string {
	doc := "Date: %v\nTo: %v\nFrom: %v\nSubject: %v\n\n%v\n"
	now := time.Now()

	return fmt.Sprintf(doc,
		now.Format(time.RFC1123),
		m.to,
		m.from,
		m.subject,
		m.body,
	)
}

func makeBody(cfg *config, r *http.Request) string {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := cfg.bodyTpl.Execute(buf, struct {
		Form url.Values
	}{
		Form: r.PostForm,
	})
	if err != nil {
		cfg.maillog.Errorf("Render Error: %s\nForm: %#v\nWritten: %s", err, r.PostForm, buf)
	}
	return buf.String()
}

func makeFrom(cfg *config, r *http.Request) string {
	from := concatEmail(cfg.fromEmail, cfg.fromName)
	if r.PostFormValue("email") != "" {
		from = concatEmail(r.PostFormValue("email"), r.PostFormValue("name"))
	}
	return from
}

func makeSubject(cfg *config, r *http.Request) string {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	err := cfg.subjectTpl.Execute(buf, struct {
		Form url.Values
	}{
		Form: r.PostForm,
	})
	if err != nil {
		cfg.maillog.Errorf("Render Subject Error: %s\nForm: %#v\nWritten: %s", err, r.PostForm, buf)
	}
	return buf.String()
}

func makeTo(cfg *config) string {
	return strings.Join(cfg.to, ", ")
}

func concatEmail(email, name string) string {
	if name != "" {
		return name + " <" + email + ">"
	}
	return email
}
