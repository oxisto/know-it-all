package teamspeak

import (
	"github.com/Darfk/ts3"
	"fmt"
	"github.com/oxisto/know-it-all/bot"
	"github.com/nlopes/slack"
	"strconv"
)

var tsClient *ts3.Client

var users map[int]string

func Init(address string, username string, password string) {
	var err error

	users = make(map[int]string)

	tsClient, err = ts3.NewClient(address)
	if err != nil {
		fmt.Printf("Could not establish connection to TS server %s. TS3 functionality not available: %v\n", address, err)
		return
	}

	_, err = tsClient.Exec(ts3.Login(username, password))
	if err != nil {
		fmt.Printf("Could not login to TS server: %v\n", err)
		return
	}
}

func ListenForEvents() {
	tsClient.NotifyHandler(NotifyHandler)
	tsClient.ExecString("use 1")
	tsClient.ExecString("servernotifyregister event=server")
}

func NotifyHandler(n ts3.Notification) {
	fmt.Printf("%v\n", n)

	if n.Type == "notifycliententerview" {
		nickname := n.Params[0]["client_nickname"]
		clientID, err := strconv.Atoi(n.Params[0]["clid"])
		if err != nil {
			fmt.Printf("Could not parse clid %s: %v\n", clientID, err)
			return
		}

		// send to Slack
		params := slack.PostMessageParameters{}
		params.AsUser = true

		// TODO: get channel id from somewhere
		bot.SendMessage("#general", fmt.Sprintf("%s joined TS3", nickname), params)

		// add to users
		users[clientID] = nickname
	} else if n.Type == "notifyclientleftview" {
		clientID, err := strconv.Atoi(n.Params[0]["clid"])
		if err != nil {
			fmt.Printf("Could not parse clid %s: %v\n", clientID, err)
			return
		}

		nickname := users[clientID]
		if nickname != "" {
			// send to Slack
			params := slack.PostMessageParameters{}
			params.AsUser = true

			bot.SendMessage("#general", fmt.Sprintf("%s left TS3", nickname), params)
		}
	}
}