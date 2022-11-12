package alertforwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/specklesystems/alertmanager-discord/pkg/alertmanager"
	"github.com/specklesystems/alertmanager-discord/pkg/discord"
	"github.com/specklesystems/alertmanager-discord/pkg/prometheus"
)

type AlertForwarder struct {
	client     httpClient
	webhookURL string
}

type httpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func NewAlertForwarder(client httpClient, webhookURL string) AlertForwarder {
	return AlertForwarder{
		client:     client,
		webhookURL: webhookURL,
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

	for status, alerts := range groupedAlerts {
		DO := discord.Out{}

		RichEmbed := discord.Embed{
			Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
			Description: amo.CommonAnnotations.Summary,
			Color:       discord.ColorGrey,
			Fields:      []discord.EmbedField{},
		}

		switch status {
		case alertmanager.StatusFiring:
			RichEmbed.Color = discord.ColorRed
		case alertmanager.StatusResolved:
			RichEmbed.Color = discord.ColorGreen
		}

		if amo.CommonAnnotations.Summary != "" {
			DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
		}

		for _, alert := range alerts {
			realname := alert.Labels["instance"]
			if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
				realname = alert.Labels["exported_instance"]
			}

			RichEmbed.Fields = append(RichEmbed.Fields, discord.EmbedField{
				Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
				Value: alert.Annotations.Description,
			})
		}

		DO.Embeds = []discord.Embed{RichEmbed}

		DOD, err := json.Marshal(DO)
		if err != nil {
			log.Printf("Error encountered when marshalling object to json. We will not continue posting to Discord. Discord Out object: '%v+'", DO)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res, err := af.client.Post(af.webhookURL, "application/json", bytes.NewReader(DOD))
		if res != nil && res.Body != nil {
			defer res.Body.Close()
		}
		if err != nil {
			log.Printf("Error encountered sending POST to '%s'.", af.webhookURL)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if res.StatusCode < 200 || res.StatusCode > 399 {
			w.WriteHeader(http.StatusInternalServerError)
			if res.Body != nil {
				io.Copy(w, res.Body)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (af *AlertForwarder) sendRawPromAlertWarn() error {
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

	DOD, err := json.Marshal(DO)
	if err != nil {
		return fmt.Errorf("Error encountered when marshalling object to json. We will not continue. Discord Out object: '%v+'. Error: %w", DO, err)
	}

	_, err = af.client.Post(af.webhookURL, "application/json", bytes.NewReader(DOD))
	if err != nil {
		return fmt.Errorf("Error encountered sending POST to '%s'. Error: %w", af.webhookURL, err)
	}

	return nil
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
		if prometheus.IsAlert(b) {
			log.Printf("Detected a Prometheus Alert, and not an AlertManager alert, has been sent within the http request. This indicates a misconfiguration. Attempting to send a message to notify the Discord channel of the misconfiguration.")
			err = af.sendRawPromAlertWarn()
			if err != nil {
				log.Printf("Error in attempting to send a warning on Discord regarding Raw Prometheus Alerts. Error: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if len(b) > 1024 {
			log.Printf("Failed to unpack inbound alert request - %s...", string(b[:1023]))
		} else {
			log.Printf("Failed to unpack inbound alert request - %s", string(b))
		}

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	af.sendWebhook(&amo, w)
}
