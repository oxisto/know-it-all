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
	ticker := time.NewTicker(5 * time.Minute)
	quit := make(chan struct{})

	tsClient.NotifyHandler(NotifyHandler)
	tsClient.ExecString("use 1")
	tsClient.ExecString("servernotifyregister event=server")

	for {
		select {
		case <-ticker.C:
			// do stuff
			resp, err := tsClient.Exec(ts3.ClientList())
			if err == nil {
				log.Printf("Updating client list...\n")
				for _, client := range resp.Params {
					clientID, err := strconv.Atoi(client["clid"])
					if err != nil {
						log.Printf("Could not parse clid %s: %v\n", client["clid"], err)
						continue
					}
					users[clientID] = client["client_nickname"]
				}
				log.Printf("%v\n", resp)
			} else {
				log.Printf("An error occured while executing ts3 command: %v\n", err)
				quit <- struct{}{}
			}
		case <-quit:
			ticker.Stop()
			log.Printf("Stopping connection to TS3 server...\n")
			break
		}
	}

	log.Printf("Connection to TS server was lost.\n")
}

func NotifyHandler(n ts3.Notification) {
	log.Printf("%v\n", n)

	if n.Type == "notifycliententerview" {
		nickname := n.Params[0]["client_nickname"]
		clientID, err := strconv.Atoi(n.Params[0]["clid"])
		if err != nil {
			log.Printf("Could not parse clid %s: %v\n", clientID, err)
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
			log.Printf("Could not parse clid %s: %v\n", n.Params[0]["clid"], err)
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
