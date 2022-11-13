package alertforwarder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"

	"github.com/rs/zerolog/log"
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
	if len(amo.Alerts) < 1 {
		log.Debug().Msg("There are no alerts within this notification. There is nothing to forward to Discord. Returning early...")
		w.WriteHeader(http.StatusOK)
		return
	}

	groupedAlerts := make(map[string][]alertmanager.Alert)
	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	failedToPublishAtLeastOne := false
	for status, alerts := range groupedAlerts {
		DO := TranslateAlertManagerToDiscord(status, amo, alerts)

		res, err := af.client.PublishMessage(DO)
		if err != nil {
			err = fmt.Errorf("Error encountered when publishing message to discord: %w", err)
			log.Error().Err(err)
			failedToPublishAtLeastOne = true
			continue
		}

		if res.StatusCode < 200 || res.StatusCode > 399 {
			failedToPublishAtLeastOne = true
			continue
		}
	}

	if failedToPublishAtLeastOne {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (af *AlertForwarder) sendRawPromAlertWarn() (*http.Response, error) {

	warningMessage := `You have probably misconfigured this software.
We detected input in Prometheus Alert format but are expecting AlertManager format.
This program is intended to ingest alerts from alertmanager.
It is not a replacement for alertmanager, it is a
webhook target for it. Please read the README.md
for guidance on how to configure it for alertmanager
or https://prometheus.io/docs/alerting/latest/configuration/#webhook_config`
	log.Warn().Msg(warningMessage)
	DO := discord.Out{
		Content: "",
		Embeds: []discord.Embed{
			{
				Title:       "You have misconfigured this software",
				Description: warningMessage,
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
	log.Info().
		Str("Host", r.Host).
		Str("Method", r.Method).
		Str("Path", r.URL.Path).
		Msg("Request received.")

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Unable to read request body.")
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
		log.Info().Msg("Detected a Prometheus Alert, and not an AlertManager alert, has been sent within the http request. This indicates a misconfiguration. Attempting to send a message to notify the Discord channel of the misconfiguration.")
		res, err := af.sendRawPromAlertWarn()
		if err != nil || (res != nil && res.StatusCode < 200 || res.StatusCode > 399) {
			statusCode := 0
			if res != nil {
				statusCode = res.StatusCode
			}

			log.Error().Err(err).Int("StatusCode", statusCode).Msg("Error when attempting to send a warning message to Discord.")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if len(b) > maxLogLength-3 {
		log.Warn().Msgf("Failed to unpack inbound alert request - %s...", string(b[:maxLogLength-3]))
	} else {
		log.Warn().Msgf("Failed to unpack inbound alert request - %s", string(b))
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}
