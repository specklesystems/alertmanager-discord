package main

import (
	"flag"
	"log"
	"os"

	"github.com/specklesystems/alertmanager-discord/pkg/server"
)


var (
	webhookURL    = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
)

func main() {
	flag.Parse()
	amds := server.AlertManagerDiscordServer{}
	stopCh, err := amds.ListenAndServe(*webhookURL, *listenAddress)
	defer func() {
		if err = amds.Shutdown(); err != nil {
			log.Fatalf("Error while shutting down server. %s", err)
		}
	}()
	if err != nil {
		close(stopCh)
	}

	// Waits here for SIGINT (kill -2) or for channel to be closed (which can occur if there is an error in the server)
	<-stopCh
}
