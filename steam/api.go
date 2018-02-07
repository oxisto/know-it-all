package steam

import (
	"github.com/oxisto/go-steamwebapi"
	"github.com/nlopes/slack"
	"github.com/StefanSchroeder/Golang-Roman"
	"time"
	"log"
	"fmt"
	"github.com/oxisto/know-it-all/bot"
)

var currentPlayers []steamwebapi.Player
var apiKey string

var games map[string]steamwebapi.Game
var apps map[string]steamwebapi.AppData

func Init(key string) {
	apiKey = key

	games = make(map[string]steamwebapi.Game)
	apps = make(map[string]steamwebapi.AppData)
}

func WatchForPlayers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C

		response, err := steamwebapi.GetPlayerSummaries(apiKey,
			[]string{"76561197962272442",
				"76561197966228499",
				"76561197960616970",
				"76561197960824521"})

		if err != nil {
			log.Printf("Error while fetching player summaries from Steam: %s\n", err.Error())
			continue
		}

		players := response.Response.Players

		if currentPlayers == nil {
			currentPlayers = players
		}

		//log.Printf("Retrieved %d player(s) from steam\n", len(currentPlayers))

		// check for differences
		for _, currentPlayer := range currentPlayers {
			// look for the player in the response
			for _, player := range players {
				if player.SteamId == currentPlayer.SteamId {
					// found the player, let's check if the game id changed
					if player.GameId != currentPlayer.GameId {
						log.Printf("Detected change for player %s\n", player.PersonaName)
						if player.GameId == "" {
							OnPlayerStoppedGame(currentPlayer)
						} else {
							OnPlayerStartedGame(player)
						}
					}
				}
			}
		}

		currentPlayers = players
	}
}

func OnPlayerStartedGame(player steamwebapi.Player) {
	// send to Slack
	params := slack.PostMessageParameters{}
	params.AsUser = false
	params.Username = player.PersonaName
	params.IconURL = player.Avatar
	params.Markdown = true
	//params.Attachments = []slack.Attachment{}
	//params.IconEmoji = ":video_game:"

	// try to find the game
	game, app := GetGame(player.GameId)

	text := fmt.Sprintf("%s is now playing *%s* (%s). Genre: %s", player.PersonaName, app.Name, app.WebSite, app.GetGenre())

	if resp, err := steamwebapi.GetPlayerAchievements(apiKey, player.GameId, player.SteamId); err == nil {
		// is there an unlocked achievement?
		if unlockedAchievement := resp.GetLatestUnlockedAchievement(); unlockedAchievement.UnlockTime != 0 {
			// try to find it
			if achievement := game.FindAchievement(unlockedAchievement.ApiName); achievement != nil {
				params.Attachments = []slack.Attachment{
					/*{
						Color:    "#00adee",
						ImageURL: app.HeaderImage,
					},*/
					{
						Color:    "#00adee",
						Title:    fmt.Sprintf("Last Unlocked Achievement: %s", achievement.DisplayName),
						Text:     achievement.Description,
						ImageURL: achievement.Icon},}
			}
		}
	}

	// find achievements
	if resp, err := steamwebapi.GetUserStatsForGame(apiKey, player.GameId, player.SteamId); err == nil {
		// special case for PAYDAY 2
		if player.GameId == "218620" {
			playerLevel := resp.GetStat("player_level")
			tankKillsGreen := resp.GetStat("enemy_kills_tank_green")
			tankKillsBlack := resp.GetStat("enemy_kills_tank_black")
			infamyLevel := resp.GetNumStatsWithPrefix("player_rank_") - 1

			attachment := slack.Attachment{
				Color: "#00adee",
				Title: "PAYDAY 2 Stats",
				Text:  fmt.Sprintf("Player Level: %v-%d\nBulldoozer (green) Kills: %d\nBulldoozer (black) Kills: %d", roman.Roman(infamyLevel), playerLevel.Value, tankKillsGreen.Value, tankKillsBlack.Value)}

			params.Attachments = append(params.Attachments, attachment)
		}
	}

	log.Printf("%s\n", text)

	bot.SendMessage("#general", text, params)
}

func OnPlayerStoppedGame(player steamwebapi.Player) {
	// send to Slack
	params := slack.PostMessageParameters{}
	params.AsUser = false
	params.Username = player.PersonaName
	params.IconURL = player.Avatar

	// try to find the game
	_, app := GetGame(player.GameId)

	text := fmt.Sprintf("%s stopped playing %s.", player.PersonaName, app.Name)

	log.Printf("%s\n", text)

	bot.SendMessage("#general", text, params)
}

func GetGame(appID string) (steamwebapi.Game, steamwebapi.AppData) {
	if _, exists := games[appID]; !exists {
		log.Printf("Trying to fetch game %s from Steam\n", appID)
		// try to fetch game info
		if response, err := steamwebapi.GetSchemaForGame(apiKey, appID); err == nil {
			games[appID] = response.Game
		}

		if response2, err := steamwebapi.GetAppDetails([]string{appID}); err == nil {
			apps[appID] = response2[appID].Data
		}
	}

	return games[appID], apps[appID]
}
