package alertforwarder

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/specklesystems/alertmanager-discord/pkg/discord"
)

func Test_TransformAndForward_HappyPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{
      "alerts":[{"status":"firing"}],
      "commonAnnotations":{"summary":""},
      "commonLabels":{"alertname":""},
      "externalURL":"",
      "groupKey":"",
      "groupLabels":{"alertname":""},
      "receiver":"",
      "status":"",
      "version":""}`))
	req.Host = "testing.localhost"

	mockClientRecorder := mockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code to be %d, got %d", http.StatusOK, res.StatusCode)
	}
	if !mockClientRecorder.ClientTriggered {
		t.Errorf("expected mock client to have been triggered")
	}
	expectedContentType := "application/json"
	if mockClientRecorder.ContentType != expectedContentType {
		t.Errorf("expected request content type to be %s, got %s", expectedContentType, mockClientRecorder.ContentType)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(mockClientRecorder.Body)

	do := discord.Out{}
	err := json.Unmarshal(buf.Bytes(), &do)
	if err != nil {
		t.Errorf("Unexpected error marshalling to Discord Object from the Discord client request body.")
	}

	if len(do.Embeds) != 1 {
		t.Errorf("expected do.Embeds to have a length of 1, but was %d.", len(do.Embeds))
	}

	expectedColor := 10038562
	if do.Embeds[0].Color != expectedColor {
		t.Errorf("expected do.Embeds[0].Color to be %d, but was %d.", expectedColor, do.Embeds[0].Color)
	}
}

// a notification with no alerts will not be forwarded to Discord

// send a raw prom alert, receive a warning

//
