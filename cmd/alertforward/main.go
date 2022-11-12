package main

import (
	"flag"
	"os"

	"github.com/specklesystems/alertmanager-discord/pkg/server"
)


var (
	webhookURL    = flag.String("webhook.url", os.Getenv("DISCORD_WEBHOOK"), "Discord WebHook URL.")
	listenAddress = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address:Port to listen on.")
)

func main() {
	flag.Parse()
	server.Serve(*webhookURL, *listenAddress)
}
