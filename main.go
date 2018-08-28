package main

import (
	"strings"

	"os"

	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/oxisto/know-it-all/bot"
	"github.com/oxisto/know-it-all/google"
	"github.com/oxisto/know-it-all/rest"
	"github.com/oxisto/know-it-all/steam"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	SlackApiToken           = "slack-api-token"
	SlackDirectMessagesOnly = "slack-dm-only"
	GoogleApiKey            = "google-api-key"
	SteamApiKey             = "steam-api-key"
	TwitchApiKey            = "twitch-api-key"
	ListenFlag              = "listen"
	Ts3ServerFlag           = "ts3-server"
	Ts3Username             = "ts3-username"
	Ts3Password             = "ts3-password"

	DefaultListen                  = ":4300"
	DefaultSlackDirectMessagesOnly = false
)

var botCmd = &cobra.Command{
	Use:   "know-it-all",
	Short: "KnowItAll is a knowledgeable Slack bot",
	Long:  "KnowItAll is a knowledgeable Slack bot. It it almost annoying.",
	Run:   doCmd,
}

func init() {
	cobra.OnInitialize(initConfig)

	botCmd.Flags().String(ListenFlag, DefaultListen, "Host and port to listen to")
	botCmd.Flags().String(SlackApiToken, "", "The token for Slack integration")
	botCmd.Flags().String(GoogleApiKey, "", "The Google API key")
	botCmd.Flags().String(SteamApiKey, "", "The Steam API key")
	botCmd.Flags().String(TwitchApiKey, "", "The twitch API key")
	botCmd.Flags().Bool(SlackDirectMessagesOnly, DefaultSlackDirectMessagesOnly, "Should the bot interact with direct messages only?")
	botCmd.Flags().String(Ts3ServerFlag, "", "The TS3 server")
	botCmd.Flags().String(Ts3Username, "", "The TS3 username")
	botCmd.Flags().String(Ts3Password, "", "The TS3 password")
	viper.BindPFlag(ListenFlag, botCmd.Flags().Lookup(ListenFlag))
	viper.BindPFlag(SlackApiToken, botCmd.Flags().Lookup(SlackApiToken))
	viper.BindPFlag(GoogleApiKey, botCmd.Flags().Lookup(GoogleApiKey))
	viper.BindPFlag(SteamApiKey, botCmd.Flags().Lookup(SteamApiKey))
	viper.BindPFlag(TwitchApiKey, botCmd.Flags().Lookup(TwitchApiKey))
	viper.BindPFlag(SlackDirectMessagesOnly, botCmd.Flags().Lookup(SlackDirectMessagesOnly))
	viper.BindPFlag(Ts3ServerFlag, botCmd.Flags().Lookup(Ts3ServerFlag))
	viper.BindPFlag(Ts3Username, botCmd.Flags().Lookup(Ts3Username))
	viper.BindPFlag(Ts3Password, botCmd.Flags().Lookup(Ts3Password))
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func doCmd(cmd *cobra.Command, args []string) {
	log.Println("Starting bot...")

	google.InitAPI(viper.GetString(GoogleApiKey))

	//if viper.GetString(Ts3ServerFlag) != "" {
	//	teamspeak.Init(viper.GetString(Ts3ServerFlag), viper.GetString(Ts3Username), viper.GetString(Ts3Password))
	//	go teamspeak.ListenForEvents()
	//}

	steam.Init(viper.GetString(SteamApiKey))
	go steam.WatchForPlayers()

	//twitch.Init(viper.GetString(TwitchApiKey))
	//go twitch.WatchForPlayers()

	go bot.InitBot(viper.GetString(SlackApiToken), viper.GetBool(SlackDirectMessagesOnly))

	router := handlers.LoggingHandler(os.Stdout, rest.NewRouter())
	err := http.ListenAndServe(viper.GetString(ListenFlag), router)

	log.Printf("An error occured while starting the HTTP server: %s\n", err)
}

func main() {
	log.SetOutput(os.Stdout)

	if err := botCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
