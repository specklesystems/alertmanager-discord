package alertforwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"
	. "github.com/specklesystems/alertmanager-discord/test"
)

func Test_TransformAndForward_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK, nil)
	defer res.Body.Close()

	EqualInt(t, http.StatusOK, res.StatusCode, "http response status code")

	IsTrue(t, mockClientRecorder.ClientTriggered, "Should have sent a request to Discord")
	EqualStr(t, "application/json", mockClientRecorder.ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Body)
	EqualInt(t, 1, len(do.Embeds), "Discord message embed length")
	EqualInt(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	Contains(t, "a_common_annotation_summary", do.Content, "Discord message content")
}

func Test_TransformAndForward_InvalidInput_NoValue_ReturnsErrorResponseCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest, nil)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	EqualInt(t, http.StatusBadRequest, res.StatusCode, "Should expect an http response status code indicating request was bad.")

	IsFalse(t, mockClientRecorder.ClientTriggered, "should not have sent a request to Discord")
}

func Test_TransformAndForward_InvalidInput_LongString_ReturnsErrorResponseCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 1025)))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest, nil)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	EqualInt(t, http.StatusBadRequest, res.StatusCode, "Should expect an http response status code indicating request was bad.")

	IsFalse(t, mockClientRecorder.ClientTriggered, "should not have sent a request to Discord")
}

func Test_TransformAndForward_InvalidInput_PrometheusAlert_ReturnsErrorResponseCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest, nil)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	EqualInt(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating server internal error.")

	IsTrue(t, mockClientRecorder.ClientTriggered, "should have sent a request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

func Test_TransformAndForward_PrometheusAlert_And_DiscordClientResponsdsWithError_RespondsWithErrorCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest, fmt.Errorf("Discord client responds with an error."))

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	EqualInt(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating request was unprocessable.")

	IsTrue(t, mockClientRecorder.ClientTriggered, "should have sent a request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

func Test_TransformAndForward_PrometheusAlert_And_DiscordClientResponsdsWithErrorStatusCode_RespondsWithErrorStatusCode(t *testing.T) {
	promAlert := []prometheus.Alert{
		{
			Status: "",
		},
	}
	promAlertJson, err := json.Marshal(promAlert)
	NoError(t, err, "marshalling prometheus alert")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(promAlertJson))
	req.Host = "testing.localhost"

	mockClientRecorder := MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(http.StatusBadRequest, nil)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	res := w.Result()
	defer res.Body.Close()

	EqualInt(t, http.StatusInternalServerError, res.StatusCode, "Should expect an http response status code indicating internal server error.")

	IsTrue(t, mockClientRecorder.ClientTriggered, "should have sent a request to Discord (with a message stating there is a problem)")
	// TODO test message content sent to Discord
}

func Test_TransformAndForward_NoAlerts_DoesNotSendToDiscord(t *testing.T) {
	ao := alertmanager.Out{}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusBadRequest, nil)
	defer res.Body.Close()

	EqualInt(t, http.StatusOK, res.StatusCode, "http response status code")

	IsFalse(t, mockClientRecorder.ClientTriggered, "mock client should not be triggered")
}

func Test_TransformAndForward_NoCommonAnnotationSummary_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK, nil)
	defer res.Body.Close()

	EqualInt(t, http.StatusOK, res.StatusCode, "http response status code")

	IsTrue(t, mockClientRecorder.ClientTriggered, "mock client should be triggered")
	EqualStr(t, "application/json", mockClientRecorder.ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Body)
	EqualInt(t, 1, len(do.Embeds), "Discord message embed length")
	EqualInt(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	EqualStr(t, "", do.Content, "Discord message content")
}

func Test_TransformAndForward_StatusResolved_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusResolved,
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK, nil)
	defer res.Body.Close()

	EqualInt(t, http.StatusOK, res.StatusCode, "http response status code")

	IsTrue(t, mockClientRecorder.ClientTriggered, "mock client should be triggered")

	do := readerToDiscordOut(t, mockClientRecorder.Body)
	EqualInt(t, 1, len(do.Embeds), "Discord message embed length")
	EqualInt(t, 3066993, do.Embeds[0].Color, "Discord message embed color")
}

// alert with a label 'instance'='localhost' and 'exported_instance' label is set, should have the instance replaced by 'exported_instance'
func Test_TransformAndForward_ExportedInstance_HappyPath(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
				Labels: map[string]string{
					"instance":          "localhost",
					"exported_instance": "exported_instance_value",
				},
			},
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK, nil)
	defer res.Body.Close()

	EqualInt(t, http.StatusOK, res.StatusCode, "http response status code")

	IsTrue(t, mockClientRecorder.ClientTriggered, "mock client should be triggered")
	EqualStr(t, "application/json", mockClientRecorder.ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Body)
	EqualInt(t, 1, len(do.Embeds), "Discord message embed length")
	EqualInt(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	EqualInt(t, 1, len(do.Embeds[0].Fields), "Discord message embed fields length")
	Contains(t, "exported_instance_value", do.Embeds[0].Fields[0].Name, "Discord message embed field Name should contain instance")
	EqualStr(t, "", do.Content, "Discord message content")
}

// Discord client returns an error (e.g. a closed connection, network outage or similar)
func Test_TransformAndForward_DiscordClientReturnsError(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusOK, fmt.Errorf("an error in the Discord client."))
	defer res.Body.Close()

	EqualInt(t, http.StatusInternalServerError, res.StatusCode, "http response status code")

	IsTrue(t, mockClientRecorder.ClientTriggered, "Should have sent a request to Discord")
	EqualStr(t, "application/json", mockClientRecorder.ContentType, "content type")

	do := readerToDiscordOut(t, mockClientRecorder.Body)
	EqualInt(t, 1, len(do.Embeds), "Discord message embed length")
	EqualInt(t, 10038562, do.Embeds[0].Color, "Discord message embed color")
	Contains(t, "a_common_annotation_summary", do.Content, "Discord message content")
}

func Test_TransformAndForward_DiscordReturnsWithErrorStatusCode_ReturnInternalServerErrorStatusCode(t *testing.T) {
	ao := alertmanager.Out{
		Alerts: []alertmanager.Alert{
			{
				Status: alertmanager.StatusFiring,
			},
		},
		CommonAnnotations: struct {
			Summary string `json:"summary"`
		}{
			Summary: "a_common_annotation_summary",
		},
	}

	mockClientRecorder, res := triggerAndRecordRequest(t, ao, http.StatusUnauthorized, nil)
	defer res.Body.Close()

	IsTrue(t, mockClientRecorder.ClientTriggered, "Should have sent a request to Discord")
	EqualStr(t, "application/json", mockClientRecorder.ContentType, "content type")

	EqualInt(t, http.StatusInternalServerError, res.StatusCode, "http response status code should be 500")
}

// TODO Add a test for context with multiple alerts: if some are firing and some resolved we should publish two separate messages to Discord - alerts with matching statuses should be grouped together

// HELPERS

func triggerAndRecordRequest(t *testing.T, request alertmanager.Out, discordStatusCode int, discordClientError error) (mockClientRecorder MockClientRecorder, httpResponse *http.Response) {
	aoJson, err := json.Marshal(request)
	NoError(t, err, "marshalling alertmanager out")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(aoJson))
	req.Host = "testing.localhost"

	mockClientRecorder = MockClientRecorder{}
	mockClient := mockClientRecorder.NewMockClientWithResponse(discordStatusCode, discordClientError)

	SUT := NewAlertForwarder(mockClient, "https://discordapp.com/api/webhooks/123456789123456789/abc")

	w := httptest.NewRecorder()
	SUT.TransformAndForward(w, req)

	httpResponse = w.Result()
	return mockClientRecorder, httpResponse
}

func readerToDiscordOut(t *testing.T, reader io.Reader) discord.Out {
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	do := discord.Out{}
	err := json.Unmarshal(buf.Bytes(), &do)
	if err != nil {
		t.Errorf("Unexpected error marshalling to Discord Object from the Discord client request body.")
	}
	return do
}
