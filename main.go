package main

import (
	"fmt"
	"strings"

	"os"

	"github.com/oxisto/know-it-all/bot"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	SlackApiToken           = "slack-api-token"
	SlackDirectMessagesOnly = "slack-dm-only"
	GoogleApiKey            = "google-api-key"

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

	botCmd.Flags().String(SlackApiToken, "", "The token for Slack integration")
	botCmd.Flags().String(GoogleApiKey, "", "The Google API key")
	botCmd.Flags().Bool(SlackDirectMessagesOnly, DefaultSlackDirectMessagesOnly, "Should the bot interact with direct messages only?")
	viper.BindPFlag(SlackApiToken, botCmd.Flags().Lookup(SlackApiToken))
	viper.BindPFlag(GoogleApiKey, botCmd.Flags().Lookup(GoogleApiKey))
	viper.BindPFlag(SlackDirectMessagesOnly, botCmd.Flags().Lookup(SlackDirectMessagesOnly))
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}

func doCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Starting bot...")

	bot.InitGoogleAPI(viper.GetString(GoogleApiKey))

	bot.InitBot(viper.GetString(SlackApiToken),
		viper.GetBool(SlackDirectMessagesOnly),
		viper.GetString(GoogleApiKey))
}

func main() {
	if err := botCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
