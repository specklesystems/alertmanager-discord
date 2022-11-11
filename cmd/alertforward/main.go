package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/benjojo/alertmanager-discord/pkg/alertforwarder"
)

const defaultListenAddress = "127.0.0.1:9094"

var (
	whURL         = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
)

const (
	FaviconPath   = "/favicon.ico"
	LivenessPath  = "/liveness"
	ReadinessPath = "/readiness"
)

func main() {
	flag.Parse()
	alertforwarder.CheckWhURL(*whURL)

	if *listenAddress == "" {
		*listenAddress = defaultListenAddress
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	af := alertforwarder.NewAlertForwarder(client, *whURL)

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

	log.Printf("Listening on: %s", *listenAddress)

	log.Fatalf("Failed to listen on HTTP: %v",
		http.ListenAndServe(*listenAddress, nil))
}
