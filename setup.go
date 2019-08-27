package mailgun

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/rojakcoder/mailgun-caddy/maillog"
)

func init() {
	caddy.RegisterPlugin("mailgun", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
	// This line is only for a plugin under development. When the plugin is formally recognized and placed into plugin.go - `directives`, this line will not be needed.
	httpserver.RegisterDevDirective("mailgun", "")
	fmt.Println("> Initiating mailgun-caddy")
}

// setup used internally by Caddy to set up this middleware
func setup(c *caddy.Controller) error {
	mc, err := parse(c)
	if err != nil {
		return err
	}

	if c.ServerBlockKeyIndex == 0 {
		// only run when the first hostname has been loaded.
		if mc.maillog, err = mc.maillog.Init(c.ServerBlockKeys...); err != nil {
			return err
		}
		if err = mc.loadFromEnv(); err != nil {
			return err
		}
		if err = mc.loadTemplate(); err != nil {
			return err
		}

		c.ServerBlockStorage = newHandler(mc)
	}

	if moh, ok := c.ServerBlockStorage.(*handler); ok { // moh = mailOutHandler ;-)
		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			moh.Next = next
			return moh
		})
		return nil
	}
	return errors.New("[mailgun] Failed to create middleware handler")
}

func parse(c *caddy.Controller) (mc *config, _ error) {
	// This parses the following config blocks
	mc = newConfig()

	for c.Next() {
		args := c.RemainingArgs()

		switch len(args) {
		case 1:
			mc.endpoint = args[0]
		}

		for c.NextBlock() {
			var err error
			switch c.Val() {
			case "domain":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.domain = c.Val()
			case "privatekey":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.privatekey = c.Val()
			case "maillog":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				if mc.maillog.IsNil() {
					mc.maillog = maillog.New(c.Val(), "")
				} else {
					mc.maillog.MailDir = c.Val()
				}
			case "errorlog":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				if mc.maillog.IsNil() {
					mc.maillog = maillog.New("", c.Val())
				} else {
					mc.maillog.ErrDir = c.Val()
				}
			case "from_email":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.fromEmail = c.Val()
			case "from_name":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.fromName = c.Val()
			case "to":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.to, err = splitEmails(c.Val())
				if err != nil {
					return nil, err
				}
			case "cc":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.cc, err = splitEmails(c.Val())
				if err != nil {
					return nil, err
				}
			case "bcc":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.bcc, err = splitEmails(c.Val())
				if err != nil {
					return nil, err
				}
			case "subject":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.subject = c.Val()
			case "body":
				if !c.NextArg() {
					return nil, c.ArgErr()
				}
				mc.body = c.Val()
			}
		}
	}
	missings := make([]string, 0, 2)
	if mc.domain == "" {
		missings = append(missings, "domain")
	}
	if mc.privatekey == "" {
		missings = append(missings, "privatekey")
	}
	if len(missings) > 0 {
		msg := "The following %v required and cannot be empty: %v"
		if len(missings) > 1 {
			return mc, fmt.Errorf(msg, "properties are", strings.Join(missings, ", "))
		} else {
			return mc, fmt.Errorf(msg, "property is", strings.Join(missings, ", "))
		}
	}
	return
}
