package cmd

import (
	"os"
	"time"

	"github.com/specklesystems/alertmanager-discord/pkg/server"
	"github.com/specklesystems/alertmanager-discord/pkg/version"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	discordWebhookUrlFlagKey       = "discord.webhook.url"
	discordWebhookUrlEnvVarKey     = "DISCORD_WEBHOOK_URL"
	listenAddressFlagKey           = "listen.address"
	listenAddressEnvVarKey         = "LISTEN_ADDRESS"
	maxBackoffTimeSecondsFlagKey   = "max.backoff.time.seconds"
	maxBackoffTimeSecondsEnvVarKey = "MAX_BACKOFF_TIME_SECONDS"
)

var (
	webhookURL                string
	listenAddress             string
	maximumBackoffTimeSeconds int
)

func init() {
	rootCmd.Flags().StringVarP(&webhookURL, discordWebhookUrlFlagKey, "d", "", "Url to the Discord webhook API endpoint.")
	viper.BindPFlag(discordWebhookUrlFlagKey, rootCmd.Flags().Lookup(discordWebhookUrlFlagKey))
	viper.BindEnv(discordWebhookUrlFlagKey, discordWebhookUrlEnvVarKey)

	rootCmd.Flags().StringVarP(&listenAddress, listenAddressFlagKey, "l", server.DefaultListenAddress, "The address (host:port) which the server will attempt to bind to and listen on.")
	viper.BindPFlag(listenAddressFlagKey, rootCmd.Flags().Lookup(listenAddressFlagKey))
	viper.BindEnv(listenAddressFlagKey, listenAddressEnvVarKey)

	rootCmd.Flags().IntVarP(&maximumBackoffTimeSeconds, maxBackoffTimeSecondsFlagKey, "", 10, "The maximum elapsed duration (expressed as an integer number of seconds) to allow the Discord client to continue retrying to send messages to the Discord API.")
	viper.BindPFlag(maxBackoffTimeSecondsFlagKey, rootCmd.Flags().Lookup(maxBackoffTimeSecondsFlagKey))
	viper.BindEnv(maxBackoffTimeSecondsFlagKey, maxBackoffTimeSecondsEnvVarKey)
}

var rootCmd = &cobra.Command{
	Use:     "alertmanager-discord",
	Version: version.Version,
	Short:   "Forwards AlertManager alerts to Discord.",
	Long: `A simple web server that accepts AlertManager webhooks,
translates the data to match Discord's message specifications,
and forwards that to Discord's message API endpoint.`,
	Run: func(cmd *cobra.Command, args []string) {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		amds := server.AlertManagerDiscordServer{
			MaximumBackoffTimeSeconds: time.Duration(maximumBackoffTimeSeconds) * time.Second,
		}
		stopCh, err := amds.ListenAndServe(webhookURL, listenAddress)
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
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Error when executing command. Exiting program...")
		os.Exit(1)
	}
}
