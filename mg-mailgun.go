package mailgun

import (
	"context"
	"fmt"
	"net/http"
	"time"

	mg "github.com/mailgun/mailgun-go/v3"
)

// SendEmail is the function that sends the request to the Mailgun API to deliver the email.
//
// The message is logged to the maillog if available.
func SendEmail(cfg *config, r *http.Request) error {
	svc := mg.NewMailgun(cfg.domain, cfg.privatekey)

	msg := newMessage(cfg, r)

	svcMsg := svc.NewMessage(msg.from, msg.subject, msg.body, cfg.to...)
	if cfg.bodyIsHTML {
		svcMsg.SetHtml(msg.body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := svc.Send(ctx, svcMsg)
	if err != nil {
		cfg.maillog.Errorf("Failed to send email: %v\n", err)
		return err
	}
	// Log the message
	wout := cfg.maillog.NewWriter()
	fmt.Fprintf(wout, msg.copy())
	return nil
}
