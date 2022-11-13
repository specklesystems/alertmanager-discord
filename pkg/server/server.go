package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/alertforwarder"

	"github.com/rs/zerolog/log"
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

func (amds *AlertManagerDiscordServer) ListenAndServe(webhookUrl, listenAddress string) (chan os.Signal, error) {
	stop := make(chan os.Signal, 1)
	mux := http.NewServeMux()

	ok, _, err := alertforwarder.CheckWebhookURL(webhookUrl)
	if !ok {
		return stop, fmt.Errorf("url is invalid: %w", err)
	}

	if listenAddress == "" {
		log.Info().Msgf("Listen address not provided. Using default: '%s'", defaultListenAddress)
		listenAddress = defaultListenAddress
	}
	log.Info().Msgf("Listening on: %s", listenAddress)

	discordClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	af := alertforwarder.NewAlertForwarder(discordClient, webhookUrl)

	mux.HandleFunc("/", af.TransformAndForward)

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Readiness probe encountered.")
	})

	mux.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		log.Info().Msg("Liveness probe encountered.")
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		// purposefully empty
	})

	amds.httpServer = &http.Server{
		Addr:           listenAddress,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Setting up signal capturing
	signal.Notify(stop, os.Interrupt)

	go func() {
		// check for nil prevents race condition if we have already shutdown the server before this goroutine attempts to start
		if amds.httpServer != nil {
			if err := amds.httpServer.ListenAndServe(); err != nil {
				close(stop)
			}
		}
	}()

	return stop, nil
}

func (amds *AlertManagerDiscordServer) Shutdown() error {
	log.Info().Msg("Received signal to shut down server. Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if amds.httpServer == nil {
		// http server is not referenced, or was never created, so we're unable to shut it down
		return nil
	}

	if err := amds.httpServer.Shutdown(ctx); err != nil {
		// prevent race condition if shutdown signal was sent prior to server starting, we remove server reference to prevent it starting
		amds.httpServer = nil
		return err
	}

	// prevent race condition if shutdown signal was sent prior to server starting, we remove server remove to prevent it starting
	amds.httpServer = nil
	return nil
}
