package server

import (
	"log"
	"net/http"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertforwarder"
)

const defaultListenAddress = "127.0.0.1:9094"
const (
	FaviconPath   = "/favicon.ico"
	LivenessPath  = "/liveness"
	ReadinessPath = "/readiness"
)

func Serve(webhookUrl, listenAddress string) {
	ok, _ := alertforwarder.CheckWebhookURL(webhookUrl)
	if !ok {
		log.Fatal("URL is invalid, exiting program...")
	}

	if listenAddress == "" {
		log.Printf("Listen address not provided. Using default: '%s'", defaultListenAddress)
		listenAddress = defaultListenAddress
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	af := alertforwarder.NewAlertForwarder(client, webhookUrl)

	http.HandleFunc("/", af.TransformAndForward)

	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Readiness probe encountered.")
	})

	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Liveness probe encountered.")
	})

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// purposefully empty
	})

	log.Printf("Listening on: %s", listenAddress)

	log.Fatalf("Failed to listen on HTTP: %v",
		http.ListenAndServe(listenAddress, nil))
}
