package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	. "github.com/specklesystems/alertmanager-discord/test"
)

const (
	serverListenAddress = "127.0.0.1:9096"
)

func Test_Serve_HappyPath(t *testing.T) {
	// create a mock discord server to respond to our request
	receivedRequest := make(chan bool, 1)
	mockDiscordServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Discord mock server will always return with Status Code 200 OK
		receivedRequest <- true // notify the channel that the discord server received the request
	}))
	defer mockDiscordServer.Close()

	amds := AlertManagerDiscordServer{}
	defer func() {
		err := amds.Shutdown()
		NoError(t, err, "server shutdown should not error")
	}()

	_, err := amds.ListenAndServe(mockDiscordServer.URL, serverListenAddress)
	NoError(t, err, "server ListenAndServe should not error")

	client := http.Client{
		Timeout: 500 * time.Millisecond,
	}

	res, err := client.Get(fmt.Sprintf("http://%s/liveness", serverListenAddress))
	NotNil(t, res, "response to GET liveness should not be nil")
	EqualInt(t, http.StatusOK, res.StatusCode, "GET liveness should return status code OK (200)")
	res, err = client.Get(fmt.Sprintf("http://%s/readiness", serverListenAddress))
	NotNil(t, res, "response to GET readiness should not be nil")
	EqualInt(t, http.StatusOK, res.StatusCode, "GET readiness should return status code OK (200)")
	res, err = client.Get(fmt.Sprintf("http://%s/favicon.ico", serverListenAddress))
	NotNil(t, res, "response to GET favicon.ico should not be nil")
	EqualInt(t, http.StatusOK, res.StatusCode, "GET favicon.ico should return status code OK (200)")

	// assert mock discord server received expected json
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

	aoJson, err := json.Marshal(ao)
	NoError(t, err, "marshalling alertmanager out")

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/", serverListenAddress), bytes.NewReader(aoJson))
	NoError(t, err, "creating http request")
	req.Host = mockDiscordServer.URL

	res, err = client.Do(req)
	defer func() {
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
	}()
	NoError(t, err, "sending request to alertmanager-discord server.")
	NotNil(t, res, "response to POST '/' should not be nil")
	EqualInt(t, http.StatusOK, res.StatusCode, "sending valid alertmanager data should expect http response status code")
	IsTrue(t, <-receivedRequest, "Mock discord server should have received response") // will wait until the request is received

	// TODO assert log lines were generated
}

// Test with invalid URL, throws an error
func Test_Server_InvalidDiscordUrl(t *testing.T) {
	amds := AlertManagerDiscordServer{}
	defer func() {
		err := amds.Shutdown()
		NoError(t, err, "server shutdown should not error")
	}()

	_, err := amds.ListenAndServe("https://example.org/not/a/discord/webhook/api", "127.0.0.1:9095")
	HasError(t, err, "server ListenAndServe should return an error for an invalid url")
}

// Test that no listen address will set to default listen address
func Test_Server_With_EmptyListenAddress_DefaultsToListenAddress(t *testing.T) {
	amds := AlertManagerDiscordServer{}
	defer func() {
		err := amds.Shutdown()
		NoError(t, err, "server shutdown should not error")
	}()

	_, err := amds.ListenAndServe("http://localhost/", "")
	NoError(t, err, "server ListenAndServe should not error")

	client := http.Client{
		Timeout: 500 * time.Millisecond,
	}

	// it should have defaulted to the default listen address
	res, err := client.Get(fmt.Sprintf("http://%s/liveness", "127.0.0.1:9094"))
	NotNil(t, res, "Response should not be nil")
	EqualInt(t, http.StatusOK, res.StatusCode, "Liveness probe should return status code OK (200)")
}
