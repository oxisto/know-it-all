package bot

import (
	"fmt"

	"strings"

	"errors"

	"math/rand"

	"github.com/nlopes/slack"
)

type Something struct {
	Channel string
	Msg     slack.Msg
}

var (
	replyChannel chan Something
	api          *slack.Client
	botId        string
)

var googleApiKey string

func Bot(token string, apiKey string) {
	googleApiKey = apiKey

	fmt.Println("Connecting to Slack...")

	api = slack.New(token)

	replyChannel = make(chan Something)
	go handleBotReply()

	//api.SendMessage("#eve", slack.MsgOptionText("EVE Corpus2 Server has started.", false), slack.MsgOptionPost(), slack.MsgOptionAsUser(true))

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				botId = ev.Info.User.ID
				fmt.Printf("Connected to Slack using %s\n", botId)
			case *slack.TeamJoinEvent:
				// Handle new user to client
			case *slack.MessageEvent:
				something := Something{
					Msg:     ev.Msg,
					Channel: ev.Channel,
				}

				if ev.Msg.User != botId {
					replyChannel <- something
				}
			case *slack.ReactionAddedEvent:
				// Handle reaction added
			case *slack.ReactionRemovedEvent:
				// Handle reaction removed
			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())
			}
		}
	}
}

func handleBotReply() {
	for {
		something := <-replyChannel

		text := strings.ToLower(something.Msg.Text)

		if strings.Contains(text, "wo") {
			fmt.Println("Triggering location command...")
			locationCommand(something)
		}
	}
}

func locationCommand(something Something) {
	marks := []string{",", ".", "!", ":", "?"}
	fillerWords := map[string]bool{"liegt": true, "ist": true, "bitte": true, "danke": true, "sind": true}

	command := strings.ToLower(something.Msg.Text)

	// first, remove all punctuation marks
	for _, mark := range marks {
		command = strings.Replace(command, mark, "", -1)
	}

	// next, tokenize it and remove all filler words and re-assemble the rest after the keyword was found
	tokens := strings.Split(command, " ")
	typeNameTokens := []string{}
	keywordFound := false
	keyword := "wo"
	for _, token := range tokens {
		if token == keyword {
			keywordFound = true
			continue
		}

		if fillerWords[token] {
			continue
		}

		if keywordFound {
			typeNameTokens = append(typeNameTokens, token)
		}
	}

	locationName := strings.Join(typeNameTokens, " ")

	results, err := Geocode(locationName)
	if err != nil {
		replyWithError(something.Channel, err)
		return
	} else if len(results) == 0 {
		replyWithError(something.Channel, errors.New(fmt.Sprintf("Sorry, konnte %s nicht finden.", locationName)))
		return
	}

	// just take the first result
	result := results[0]

	// find some more details
	details, err := PlaceDetail(result.PlaceID)

	/*fields := []slack.AttachmentField{
	{
	//Value: strings.Join(details.HTMLAttributions, ""),
	}}*/

	attachment := slack.Attachment{
		Color:     "#B733FF",
		Title:     details.Name,
		TitleLink: details.URL,
		//Fields:    fields,
		ImageURL: fmt.Sprintf("https://maps.googleapis.com/maps/api/staticmap?center=%f,%f&size=640x400&zoom=9&markers=color:red|label:|%f,%f&key=%s",
			result.Geometry.Location.Lat,
			result.Geometry.Location.Lng,
			result.Geometry.Location.Lat,
			result.Geometry.Location.Lng, googleApiKey),
	}

	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.Attachments = []slack.Attachment{attachment}

	api.PostMessage(something.Channel, "", params)

	if len(details.Photos) > 0 {
		attachment = slack.Attachment{
			Color: "#B733FF",
			Title: fmt.Sprintf("Impressionen aus %s", details.Name),
			ImageURL: fmt.Sprintf("https://maps.googleapis.com/maps/api/place/photo?maxwidth=400&photoreference=%s&key=%s",
				details.Photos[rand.Int31n(int32(len(details.Photos)))].PhotoReference,
				googleApiKey),
		}

		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.Attachments = []slack.Attachment{attachment}

		api.PostMessage(something.Channel, "", params)
	}
}

func replyWithError(channel string, err error) {
	api.SendMessage(channel, slack.MsgOptionText(fmt.Sprintf("Does not compute: %s", err), false), slack.MsgOptionPost(), slack.MsgOptionAsUser(true))
}
