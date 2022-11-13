package main

import (
	"flag"
	"os"

	"github.com/specklesystems/alertmanager-discord/pkg/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	webhookURL    = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	flag.Parse()
	amds := server.AlertManagerDiscordServer{}
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
