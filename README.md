This is a Caddy plugin for sending emails using the Mailgun API.

This plugin is based on the [mailout plugin](https://caddyserver.com/docs/http.mailout) that sends emails via SMTP.

The plumbing of the plugins is so similar that it makes sense to base on mailout's architecture.

To keep the similarilty intact, the file names of the files that are identical, or nearly so, are kept the same. To keep the comparison, extra effort is made to make sure the variables are named the same as well.

Files that are unique to mailgun-caddy are named with the "mg-" prefix.

## Differences

For some reason, the rate limiting functionality is not working. The section in serve.go encounters some error resulting in a 500 response. No effort has been placed in identifying the cause of the error.

If the files with the same names do differ, they either differ in some minor aspects (e.g. "mailgun" instead of "mailout") or some function in mailout is removed as they are not needed in mailgun.  