package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertforwarder"
)

const defaultListenAddress = "127.0.0.1:9094"
const (
	FaviconPath   = "/favicon.ico"
	LivenessPath  = "/liveness"
	ReadinessPath = "/readiness"
)

type AlertManagerDiscordServer struct {
	httpServer *http.Server
}

func (amds *AlertManagerDiscordServer) ListenAndServe(webhookUrl, listenAddress string) (error, chan os.Signal) {
	mux := http.NewServeMux()

	ok, _ := alertforwarder.CheckWebhookURL(webhookUrl)
	if !ok {
		log.Fatal("URL is invalid, exiting program...")
	}

	if listenAddress == "" {
		log.Printf("Listen address not provided. Using default: '%s'", defaultListenAddress)
		listenAddress = defaultListenAddress
	}
	log.Printf("Listening on: %s", listenAddress)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	af := alertforwarder.NewAlertForwarder(client, webhookUrl)

	mux.HandleFunc("/", af.TransformAndForward)

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Readiness probe encountered.")
	})

	mux.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		log.Print("Liveness probe encountered.")
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// purposefully empty
	})

	amds.httpServer = &http.Server{
		Addr: listenAddress,
		Handler: mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
			if err := amds.httpServer.ListenAndServe(); err != nil {
				close(stop)
			}
	}()

	return nil, stop
}

func (amds *AlertManagerDiscordServer) Shutdown() error {
	log.Print("Received signal to shut down server. Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := amds.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Error received on server shutdown: %s", err)
		amds.httpServer = nil
		return err
	}

	amds.httpServer = nil
	return nil
}
