package bot

import (
	"fmt"

	"strings"

	"errors"

	"math/rand"

	"github.com/nlopes/slack"
	"github.com/oxisto/know-it-all/wikipedia"
)

type Something struct {
	Channel string
	Msg     slack.Msg
}

type State struct {
	GoogleApiKey       string
	DirectMessagesOnly bool
	BotId              string
	ReplyChannel       chan Something
}

var state *State
var api *slack.Client

func InitBot(token string, directMessagesOnly bool, apiKey string) {
	state = &State{
		GoogleApiKey:       apiKey,
		DirectMessagesOnly: directMessagesOnly,
	}

	fmt.Println("Connecting to Slack...")

	api = slack.New(token)

	state.ReplyChannel = make(chan Something)
	go handleBotReply()

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				state.BotId = ev.Info.User.ID
				fmt.Printf("Connected to Slack as %s\n", state.BotId)
			case *slack.MessageEvent:
				something := Something{
					Msg:     ev.Msg,
					Channel: ev.Channel,
				}

				if ev.Msg.User != state.BotId && (!state.DirectMessagesOnly || state.DirectMessagesOnly && strings.HasPrefix(ev.Channel, "D")) {
					state.ReplyChannel <- something
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
		something := <-state.ReplyChannel

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

	intro, err := wikipedia.FetchIntro(details.Name)
	if err != nil {
		fmt.Printf("Could not fetch intro from Wikipedia from %s: %s\n", details.Name, err)
	}

	attachment := slack.Attachment{
		Color:     "#B733FF",
		Title:     details.Name,
		TitleLink: details.URL,
		ImageURL: fmt.Sprintf("https://maps.googleapis.com/maps/api/staticmap?center=%f,%f&size=640x400&zoom=9&markers=color:red|label:|%f,%f&key=%s",
			result.Geometry.Location.Lat,
			result.Geometry.Location.Lng,
			result.Geometry.Location.Lat,
			result.Geometry.Location.Lng, state.GoogleApiKey),
	}

	if intro != "" {
		attachment.Text = fmt.Sprintf("%s <%s/%s|_Wikipedia_>", intro, "https://de.wikipedia.org/wiki", details.Name)
		attachment.MarkdownIn = []string{"text"}
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
				state.GoogleApiKey),
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
