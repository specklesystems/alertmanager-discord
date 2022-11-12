package alertforwarder

import (
	"log"
	"net"
	"net/url"
	"regexp"
)

func CheckWebhookURL(webhookURL string) (bool, *url.URL) {
	if webhookURL == "" {
		log.Printf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
		return false, &url.URL{}
	}

	parsedUrl, err := url.Parse(webhookURL)
	if err != nil {
		log.Printf("The Discord WebHook URL doesn't seem to be a valid URL: '%s'", webhookURL)
		return false, &url.URL{}
	}

	host, _, err := net.SplitHostPort(parsedUrl.Host)
	if host == "" {
		host = parsedUrl.Host
	}

	if host == "127.0.0.1" || host == "::1" || host == "localhost" {
		log.Printf("The Discord Webhook URL is a localhost URL. This is not a valid Discord URL, but we consider it acceptable as we assume we are testing.")
		return true, parsedUrl
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18,19}/[a-zA-Z0-9_-]+`)

	ok := re.Match([]byte(webhookURL));
	if !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid: '%s'", webhookURL)
	}

	return ok, parsedUrl
}
