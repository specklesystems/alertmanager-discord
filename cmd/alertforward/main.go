package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	webhookURL                = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress             = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
	maximumBackoffTimeSeconds = flag.Int("max.backoff.time.seconds", 0, "Maximum allowed time to attempt to retry publishing to Discord before failing.")
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.Parse()
	maxBackoffTimeSeconds := parseMaxBackoffTimeSeconds()

	amds := server.AlertManagerDiscordServer{
		MaximumBackoffTimeSeconds: time.Duration(maxBackoffTimeSeconds) * time.Second,
	}
	stopCh, err := amds.ListenAndServe(*webhookURL, *listenAddress)
	defer func() {
		if err = amds.Shutdown(); err != nil {
			log.Fatal().Err(err).Msg("Error while shutting down server.")
		}
	}()
	if err != nil {
		log.Error().Err(err).Msg("Error in AlertManager-Discord server")
		close(stopCh)
	}

	// Waits here for SIGINT (kill -2) or for channel to be closed (which can occur if there is an error in the server)
	<-stopCh
}

func parseMaxBackoffTimeSeconds() int {
	var err error
	maxBackoffTimeSeconds := *maximumBackoffTimeSeconds
	if *maximumBackoffTimeSeconds <= 0 {
		environmentString := os.Getenv("MAX_BACKOFF_TIME_SECONDS")
		maxBackoffTimeSeconds, err = strconv.Atoi(environmentString)
		if err != nil {
			maxBackoffTimeSeconds = 0
			log.Warn().Msgf("Unable to parse environment variable `MAX_BACKOFF_TIME_SECONDS`: %s", environmentString)
		}
	}

	return maxBackoffTimeSeconds
}
