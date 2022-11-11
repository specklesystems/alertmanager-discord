package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Discord color values
const (
	ColorRed   = 0x992D22
	ColorGreen = 0x2ECC71
	ColorGrey  = 0x95A5A6
)

const (
	StatusFiring   = "firing"
	StatusResolved = "resolved"
)

const (
	FaviconPath   = "/favicon.ico"
	LivenessPath  = "/liveness"
	ReadinessPath = "/readiness"
)

const defaultListenAddress = "127.0.0.1:9094"

type alertManAlert struct {
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	EndsAt       string            `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
	StartsAt     string            `json:"startsAt"`
	Status       string            `json:"status"`
}

type alertManOut struct {
	Alerts            []alertManAlert `json:"alerts"`
	CommonAnnotations struct {
		Summary string `json:"summary"`
	} `json:"commonAnnotations"`
	CommonLabels struct {
		Alertname string `json:"alertname"`
	} `json:"commonLabels"`
	ExternalURL string `json:"externalURL"`
	GroupKey    string `json:"groupKey"`
	GroupLabels struct {
		Alertname string `json:"alertname"`
	} `json:"groupLabels"`
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Version  string `json:"version"`
}

type discordOut struct {
	Content string         `json:"content"`
	Embeds  []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Color       int                 `json:"color"`
	Fields      []discordEmbedField `json:"fields"`
}

type discordEmbedField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AlertForwarder struct {
	client httpClient;
}

type httpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

func NewAlertForwarder(client httpClient) AlertForwarder {
	return AlertForwarder{client: client}
}

var (
	whURL         = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
)

func checkWhURL(whURL string) {
	if whURL == "" {
		log.Fatalf("Environment variable 'DISCORD_WEBHOOK' or CLI parameter 'webhook.url' not found.")
	}
	_, err := url.Parse(whURL)
	if err != nil {
		log.Fatalf("The Discord WebHook URL doesn't seem to be a valid URL.")
	}

	re := regexp.MustCompile(`https://discord(?:app)?.com/api/webhooks/[0-9]{18,19}/[a-zA-Z0-9_-]+`)
	if ok := re.Match([]byte(whURL)); !ok {
		log.Printf("The Discord WebHook URL doesn't seem to be valid.")
	}
}

func (af *AlertForwarder) sendWebhook(amo *alertManOut) {
	groupedAlerts := make(map[string][]alertManAlert)

	for _, alert := range amo.Alerts {
		groupedAlerts[alert.Status] = append(groupedAlerts[alert.Status], alert)
	}

	for status, alerts := range groupedAlerts {
		DO := discordOut{}

		RichEmbed := discordEmbed{
			Title:       fmt.Sprintf("[%s:%d] %s", strings.ToUpper(status), len(alerts), amo.CommonLabels.Alertname),
			Description: amo.CommonAnnotations.Summary,
			Color:       ColorGrey,
			Fields:      []discordEmbedField{},
		}

		switch status {
		case StatusFiring:
			RichEmbed.Color = ColorRed
		case StatusResolved:
			RichEmbed.Color = ColorGreen
		}

		if amo.CommonAnnotations.Summary != "" {
			DO.Content = fmt.Sprintf(" === %s === \n", amo.CommonAnnotations.Summary)
		}

		for _, alert := range alerts {
			realname := alert.Labels["instance"]
			if strings.Contains(realname, "localhost") && alert.Labels["exported_instance"] != "" {
				realname = alert.Labels["exported_instance"]
			}

			RichEmbed.Fields = append(RichEmbed.Fields, discordEmbedField{
				Name:  fmt.Sprintf("[%s]: %s on %s", strings.ToUpper(status), alert.Labels["alertname"], realname),
				Value: alert.Annotations.Description,
			})
		}

		DO.Embeds = []discordEmbed{RichEmbed}

		DOD, err := json.Marshal(DO)
		if err != nil {
			log.Printf("Error encountered when marshalling object to json. We will not continue posting to Discord. Discord Out object: '%v+'", DO)
			return
		}

		_, err = af.client.Post(*whURL, "application/json", bytes.NewReader(DOD))
		if err != nil {
			log.Printf("Error encountered undertaking POST to '%s'.", *whURL)
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

	DO := discordOut{
		Content: "",
		Embeds: []discordEmbed{
			{
				Title:       "You have misconfigured this software",
				Description: badString,
				Color:       ColorGrey,
				Fields:      []discordEmbedField{},
			},
		},
	}

	DOD, err := json.Marshal(DO)
	if err != nil {
		log.Printf("Error encountered when marshalling object to json. We will not continue posting to Discord. Discord Out object: '%v+'", DO)
		return
	}

	http.Post(*whURL, "application/json", bytes.NewReader(DOD))
}

func  (af *AlertForwarder) transformAndForward(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - [%s] %s", r.Host, r.Method, r.URL.Path)

	b, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	amo := alertManOut{}
	err = json.Unmarshal(b, &amo)
	if err != nil {
		if isRawPromAlert(b) {
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

func main() {
	flag.Parse()
	checkWhURL(*whURL)

	if *listenAddress == "" {
		*listenAddress = defaultListenAddress
	}

	client := &http.Client{
			Timeout: 5 * time.Second,
	}
	af := NewAlertForwarder(client)

	http.HandleFunc("/", af.transformAndForward)
	http.HandleFunc(ReadinessPath, func(w http.ResponseWriter, r *http.Request) {
		log.Print("Readiness probe encountered.")
	})
	http.HandleFunc(LivenessPath, func(w http.ResponseWriter, r *http.Request) {
		log.Print("Liveness probe encountered.")
	})
	http.HandleFunc(FaviconPath, func(w http.ResponseWriter, r *http.Request) {
		// purposefully empty
	})

	log.Printf("Listening on: %s", *listenAddress)

	log.Fatalf("Failed to listen on HTTP: %v",
		http.ListenAndServe(*listenAddress, nil))
}
