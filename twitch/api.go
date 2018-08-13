package twitch

import (
	"log"
	"time"

	twitch "github.com/Onestay/go-new-twitch"
)

var apiKey string
var client *twitch.Client

func Init(key string) {
	apiKey = key

	client = twitch.NewClient(apiKey)

	i := twitch.GetStreamsInput{
		UserLogin: []string{"biftheki", "oxisto"},
	}

	streams, err := client.GetStreams(i) /* err != nil {
		log.Printf(users[0].Login)
	}*/

	log.Printf("%v", streams)
	log.Printf("%", err)
}

func WatchForPlayers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C

		if users, err := client.GetUsersByLogin("biftheki"); err != nil {
			log.Printf(users[0].Login)
		}

	}
}
