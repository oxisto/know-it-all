package teamspeak

import (
	"fmt"
	"log"
	"strconv"
	time "time"

	"github.com/Darfk/ts3"
	"github.com/nlopes/slack"
	"github.com/oxisto/know-it-all/bot"
)

var tsClient *ts3.Client

var users map[int]string

var ticker *time.Ticker
var quit chan struct{}

func Init(address string, username string, password string) {
	var err error

	if address == "" || username == "" || password == "" {
		log.Println("Please supply ts server address, username and password!")
		return
	}

	users = make(map[int]string)

	log.Printf("Trying to connect to TS server %s...\n", address)

	tsClient, err = ts3.NewClient(address)
	if err != nil {
		log.Printf("Could not establish connection to TS server %s. TS3 functionality not available: %v\n", address, err)
		return
	}

	log.Printf("Logging into TS server as %s...\n", username)

	_, err = tsClient.Exec(ts3.Login(username, password))
	if err != nil {
		log.Printf("Could not login to TS server: %v\n", err)
		return
	}

	log.Printf("Connection to TS server established.\n")
}

func ListenForEvents() {
	// keep alive
	ticker = time.NewTicker(3 * time.Minute)
	quit = make(chan struct{})

	tsClient.NotifyHandler(NotifyHandler)
	tsClient.ConnErrorHandler(ConnErrorHandler)
	tsClient.ExecString("use 1")
	tsClient.ExecString("servernotifyregister event=server")

	if err := UpdatePlayerList(); err != nil {
		log.Printf("An error occured while executing ts3 command: %v\n", err)
		quit <- struct{}{}
	}

test:
	for {
		select {
		case <-ticker.C:
			// do stuff
			if err := UpdatePlayerList(); err != nil {
				log.Printf("An error occured while executing ts3 command: %v\n", err)
				quit <- struct{}{}
			}
		case <-quit:
			ticker.Stop()
			log.Printf("Stopping connection to TS3 server...\n")
			break test
		}
	}

	log.Printf("Connection to TS server was lost.\n")
}

func UpdatePlayerList() error {
	resp, err := tsClient.Exec(ts3.ClientList())
	if err == nil {
		log.Printf("Updating client list...\n")
		for _, client := range resp.Params {
			// don't post regular user updates to slack, only notification
			OnUserEnter(client, false)
		}
		log.Printf("Response from ts3.ClientList(): %v\n", resp)
		return nil
	}

	return err
}

func OnUserEnter(client map[string]string, postToSlack bool) {
	clientType, err := strconv.Atoi(client["client_type"])
	if err != nil {
		log.Printf("Could not parse client_type %s: %v\n", client["client_type"], err)
		return
	}

	if clientType != 0 {
		return
	}

	nickname := client["client_nickname"]
	clientID, err := strconv.Atoi(client["clid"])
	if err != nil {
		log.Printf("Could not parse clid %s: %v\n", client["clid"], err)
		return
	}

	log.Printf("Adding %s to teamspeak user list.", nickname)
	users[clientID] = nickname

	if postToSlack {
		// send to Slack
		params := slack.PostMessageParameters{}
		params.AsUser = true

		// TODO: get channel id from somewhere
		bot.SendMessage("#general", fmt.Sprintf("%s joined TS3", nickname), params)
	}
}

func OnUserLeave(client map[string]string, postToSlack bool) {
	clientID, err := strconv.Atoi(client["clid"])
	if err != nil {
		log.Printf("Could not parse clid %s: %v\n", client["clid"], err)
		return
	}

	nickname := users[clientID]
	if nickname != "" && postToSlack {
		// send to Slack
		params := slack.PostMessageParameters{}
		params.AsUser = true

		bot.SendMessage("#general", fmt.Sprintf("%s left TS3", nickname), params)
	}
}

func ConnErrorHandler(err error) {
	log.Printf("Error: %v", err)

	// stop the ticker
	quit <- struct{}{}
}

func NotifyHandler(n ts3.Notification) {
	log.Printf("%v\n", n)

	if n.Type == "notifycliententerview" {
		OnUserEnter(n.Params[0], true)
	} else if n.Type == "notifyclientleftview" {
		OnUserLeave(n.Params[0], true)
	}
}
