package steam

import (
	"github.com/oxisto/go-steamwebapi/api"
	"fmt"
	"github.com/oxisto/go-steamwebapi/structs"
	"github.com/oxisto/know-it-all/bot"
	"github.com/nlopes/slack"
	"time"
	"log"
)

var currentPlayers []structs.Player
var apiKey string

func Init(key string) {
	apiKey = key
}

func WatchForPlayers() {
	for ; ; {
		players, err := api.FetchPlayerSummaries(apiKey, []string{"76561197962272442"})

		if err != nil {
			log.Printf("Error while fetching player summaries from Steam: %s\n", err.Error())
			continue
		}

		if currentPlayers == nil {
			currentPlayers = players
		}

		log.Printf("Retrieved %d player(s) from steam\n", len(currentPlayers))

		// check for differences
		for _, currentPlayer := range currentPlayers {
			// look for the player in the response
			for _, player := range players {
				if player.SteamId == currentPlayer.SteamId {
					// found the player, let's check if the game id changed
					if player.GameId != currentPlayer.GameId {
						log.Printf("Detected change for player %s\n", player.PersonaName)
						if player.GameId == "" {
							OnPlayerStoppedGame(player)
						} else {
							OnPlayerStartedGame(player)
						}
					}
				}
			}
		}

		time.Sleep(60 * time.Second)
	}
}

func OnPlayerStartedGame(player structs.Player) {
	// send to Slack
	params := slack.PostMessageParameters{}
	params.AsUser = true

	log.Printf("%s is now playing %d\n", player.PersonaName, player.GameId)

	bot.SendMessage("oxisto", fmt.Sprintf("%s is now playing %d", player.PersonaName, player.GameId), params)
}

func OnPlayerStoppedGame(player structs.Player) {
	// send to Slack
	params := slack.PostMessageParameters{}
	params.AsUser = true

	log.Printf("%s stopped playing %d\n", player.PersonaName, player.GameId)

	bot.SendMessage("oxisto", fmt.Sprintf("%s stopped playing %d", player.PersonaName, player.GameId), params)
}
