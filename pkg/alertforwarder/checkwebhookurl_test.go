package alertforwarder

import (
	. "github.com/specklesystems/alertmanager-discord/test"
	"testing"
)

func Test_WebhookUrl_HappyPath(t *testing.T) {
	ok, _, _ := CheckWebhookURL("https://discordapp.com/api/webhooks/123456789123456789/abc")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("https://discord.com/api/webhooks/123456789123456789/abc")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://localhost/")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://127.0.0.1/")
	IsTrue(t, ok, "Should be a valid webhook url")

	ok, _, _ = CheckWebhookURL("http://::1/")
	IsTrue(t, ok, "Should be a valid webhook url")
}

func Test_WebhookUrl_EmptyUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("")
	IsFalse(t, ok, "Empty url should be identified as invalid")
	HasError(t, err, "Empty url should return an error message")
}

func Test_WebhookUrl_InvalidUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("::::::::::")
	IsFalse(t, ok, "Malformed urls should be identified as invalid")
	HasError(t, err, "Invalid url should return an error message")
}

func Test_WebhookUrl_InvalidAPIUrl_ReturnsFalse(t *testing.T) {
	ok, _, err := CheckWebhookURL("https://discordapp.com/api/webhooks/12/abc")
	IsFalse(t, ok, "Malformed Discord API urls should be identified as invalid")
	HasError(t, err, "Malformed Discord API url should return an error message")

	ok, _, err = CheckWebhookURL("https://example.org/api/webhooks/12/abc")
	IsFalse(t, ok, "Non-Discord urls should be identified as invalid")
	HasError(t, err, "Non-Discord urls should return an error message")
}
