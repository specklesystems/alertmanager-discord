package alertforwarder

import (
	"testing"
	. "github.com/specklesystems/alertmanager-discord/test"
)

func Test_WebhookUrl_HappyPath(t *testing.T) {
	ok, _ := CheckWebhookURL("https://discordapp.com/api/webhooks/123456789123456789/abc")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _ = CheckWebhookURL("https://discord.com/api/webhooks/123456789123456789/abc")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _ = CheckWebhookURL("http://localhost/")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _ = CheckWebhookURL("http://127.0.0.1/")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _ = CheckWebhookURL("http://::1/")
	IsTrue(t, ok, "Should be a valid webhook url")
}

func Test_WebhookUrl_EmptyUrl_ReturnsFalse(t *testing.T) {
	ok, _ := CheckWebhookURL("")
	IsFalse(t, ok, "Empty url should be identified as invalid")
}

func Test_WebhookUrl_InvalidUrl_ReturnsFalse(t *testing.T) {
	ok, _ := CheckWebhookURL("::::::::::")
	IsFalse(t, ok, "Malformed urls should be identified as invalid")
}

func Test_WebhookUrl_InvalidAPIUrl_ReturnsFalse(t *testing.T) {
	ok, _ := CheckWebhookURL("https://discordapp.com/api/webhooks/12/abc")
	IsFalse(t, ok, "Malformed Discord API urls should be identified as invalid")

	ok, _ = CheckWebhookURL("https://example.org/api/webhooks/12/abc")
	IsFalse(t, ok, "Non-Discord urls should be identified as invalid")
}
