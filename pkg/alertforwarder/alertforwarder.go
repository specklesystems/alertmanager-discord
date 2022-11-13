package alertforwarder

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"
)

const (
	maxLogLength = 1024
)

type AlertForwarder struct {
	client *discord.Client
}

func NewAlertForwarder(client discord.HttpClient, webhookURL string) AlertForwarder {
	return AlertForwarder{
		client: discord.NewClient(client, webhookURL),
	}
}

func (af *AlertForwarder) sendWebhook(amo *alertmanager.Out, w http.ResponseWriter) {
	groupedAlerts := make(map[string][]alertmanager.Alert)

	if len(amo.Alerts) < 1 {
		log.Printf("There are no alerts within this notification. There is nothing to forward to Discord. Returning early...")
		w.WriteHeader(http.StatusOK)
		return
	}

	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	failedToPublishAny := false

	for status, alerts := range groupedAlerts {
		DO := TranslateAlertManagerToDiscord(status, amo, alerts)

		res, err := af.client.PublishMessage(DO)
		if err != nil {
			err = fmt.Errorf("Error encountered when publishing message to discord: %w", err)
			log.Printf("%s", err)
			failedToPublishAny = true
			continue
		}

		if res.StatusCode < 200 || res.StatusCode > 399 {
			failedToPublishAny = true
			continue
		}
	}

	if failedToPublishAny {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (af *AlertForwarder) sendRawPromAlertWarn() (*http.Response, error) {
	badString := `This program is suppose to be fed by alertmanager.` + "\n" +
		`It is not a replacement for alertmanager, it is a ` + "\n" +
		`webhook target for it. Please read the README.md  ` + "\n" +
		`for guidance on how to configure it for alertmanager` + "\n" +
		`or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`

	log.Print(`/!\ -- You have misconfigured this software -- /!\`)
	log.Print(`--- --                                      -- ---`)
	log.Print(badString)

	DO := discord.Out{
		Content: "",
		Embeds: []discord.Embed{
			{
				Title:       "You have misconfigured this software",
				Description: badString,
				Color:       discord.ColorGrey,
				Fields:      []discord.EmbedField{},
			},
		},
	}

	res, err := af.client.PublishMessage(DO)
	if err != nil {
		return nil, fmt.Errorf("Error encountered when publishing message to discord: %w", err)
	}

	return res, nil
}

func (af *AlertForwarder) TransformAndForward(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - [%s] %s", r.Host, r.Method, r.URL.Path)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("Unable to read request body.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	amo := alertmanager.Out{}
	err = json.Unmarshal(b, &amo)
	if err != nil {
		af.handleInvalidInput(b, w)
		return
	}

	af.sendWebhook(&amo, w)
}

func (af *AlertForwarder) handleInvalidInput(b []byte, w http.ResponseWriter) {
	if prometheus.IsAlert(b) {
		log.Printf("Detected a Prometheus Alert, and not an AlertManager alert, has been sent within the http request. This indicates a misconfiguration. Attempting to send a message to notify the Discord channel of the misconfiguration.")
		res, err := af.sendRawPromAlertWarn()
		if err != nil || (res != nil && res.StatusCode < 200 || res.StatusCode > 399) {
			statusCode := 0
			if res != nil {
				statusCode = res.StatusCode
			}

			log.Printf("Error in attempting to send a warning on Discord regarding Raw Prometheus Alerts. Status Code: '%d', Error: %s", statusCode, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if len(b) > maxLogLength-3 {
		log.Printf("Failed to unpack inbound alert request - %s...", string(b[:maxLogLength-3]))
	} else {
		log.Printf("Failed to unpack inbound alert request - %s", string(b))
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
