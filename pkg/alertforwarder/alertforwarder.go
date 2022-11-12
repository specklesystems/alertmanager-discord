package alertforwarder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/benjojo/alertmanager-discord/pkg/alertmanager"
	"github.com/benjojo/alertmanager-discord/pkg/discord"
	"github.com/benjojo/alertmanager-discord/pkg/prometheus"
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

func CheckWebhookURL(webhookURL string) {
	if webhookURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}
	_, err := url.Parse(webhookURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18,19}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(webhookURL)); !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid.")
	}
}

func (af *AlertForwarder) sendWebhook(amo *alertmanager.Out) {
	groupedAlerts := make(map[string][]alertmanager.Alert)

	if len(amo.Alerts) < 1 {
		log.Printf("There are no alerts within this notification. There is nothing to forward to Discord. Returning early...")
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
			return
		}

		_, err = af.client.Post(af.webhookURL, "application/json", bytes.NewReader(DOD))
		if err != nil {
			log.Printf("Error encountered undertaking POST to '%s'.", af.webhookURL)
		}
	}
}

func (af *AlertForwarder) sendRawPromAlertWarn() {
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
		log.Printf("Error encountered when marshalling object to json. We will not continue. Discord Out object: '%v+'", DO)
		return
	}

	http.Post(af.webhookURL, "application/json", bytes.NewReader(DOD))
}

func (af *AlertForwarder) TransformAndForward(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - [%s] %s", r.Host, r.Method, r.URL.Path)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	amo := alertmanager.Out{}
	err = json.Unmarshal(b, &amo)
	if err != nil {
		if prometheus.IsAlert(b) {
			af.sendRawPromAlertWarn()
			return
		}

		if len(b) > 1024 {
			log.Printf("Failed to unpack inbound alert request - %s...", string(b[:1023]))
		} else {
			log.Printf("Failed to unpack inbound alert request - %s", string(b))
		}

		return
	}

	af.sendWebhook(&amo)
}
