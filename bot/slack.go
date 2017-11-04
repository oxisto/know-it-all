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

		tokens := TokenizeAndNormalize(something.Msg.Text)

		if index := tokens.ContainsWord("wo"); index != -1 {
			locationCommand(something, tokens, index)
		} else if index := tokens.ContainsWord("was"); index != -1 {
			lookupCommand(something, tokens, index)
		}
	}
}

type Tokens []string
type Token string
type InvertedIndex map[string][]int

func TokenizeAndNormalize(message string) Tokens {
	marks := []string{",", ".", "!", ":", "?"}
	fillerWords := []string{"eigentlich", "ist", "bitte", "danke", "sind", "der", "die", "das", "ein", "denn"}

	message = strings.ToLower(message)

	// first, remove all punctuation marks
	for _, mark := range marks {
		message = strings.Replace(message, mark, "", -1)
	}

	// next, tokenize it and remove all filler words
	tokens := Tokens(strings.Split(message, " "))

	// this inverted index contains all tokens as well as their position in the original sentence
	idx := tokens.BuildInvertedIndex()

	// remove filler words
	for _, word := range fillerWords {
		if positions := idx[word]; positions != nil {
			for _, position := range positions {
				// not the best, but the easiest solution, empty tokens should be ignored later on
				tokens[position] = ""
			}
		}
	}

	return tokens
}

func (tokens Tokens) BuildInvertedIndex() InvertedIndex {
	invertedIndex := InvertedIndex{}

	// build the reverse index
	for i, token := range tokens {
		// if the token already exists in the index, add the position
		if positions := invertedIndex[token]; positions != nil {
			positions = append(positions, i)
		} else {
			invertedIndex[token] = []int{i}
		}
	}

	return invertedIndex
}

func (tokens Tokens) ContainsWord(word string) (index int) {
	index = -1
	for i, w := range tokens {
		if w == word {
			index = i
			break
		}
	}

	return index
}

func (tokens Tokens) Reassemble() string {
	// get rid of extra whitespaces that originally were filler words
	return strings.Trim(strings.Replace(strings.Join(tokens, " "), "  ", " ", -1), " ")
}

func lookupCommand(something Something, tokens Tokens, index int) {
	fmt.Println("Triggering lookup command...")

	thing := tokens[index+1:].Reassemble()

	if thing == "" {
		// just ignore it
		return
	}

	fmt.Printf("Trying to lookup %s...\n", thing)

	intro, extract, err := wikipedia.FetchIntro(thing)
	if err != nil {
		fmt.Printf("Could not fetch intro from Wikipedia for %s: %s\n", thing, err)
	}

	attachment := slack.Attachment{
		Color:     "#B733FF",
		Title:     extract.Title,
		TitleLink: fmt.Sprintf("https://de.wikipedia.org/wiki/%s", extract.Title),
		Text: fmt.Sprintf("%s <%s/%s|_Wikipedia_>", intro, "https://de.wikipedia.org/wiki", extract.Title),
		MarkdownIn: []string{"text"},
	}

	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.Attachments = []slack.Attachment{attachment}

	api.PostMessage(something.Channel, "", params)
}

func locationCommand(something Something, tokens Tokens, index int) {
	fmt.Println("Triggering location command...")

	locationName := tokens[index+1:].Reassemble()

	if locationName == "" {
		// just ignore it
		return
	}

	fmt.Printf("Trying to locate %s...\n", locationName)

	results, err := Geocode(locationName)
	if err != nil {
		replyWithError(something, err)
		return
	} else if len(results) == 0 {
		replyWithError(something, errors.New(fmt.Sprintf("No results for '%s'.", locationName)))
		return
	}

	// just take the first result
	result := results[0]

	// find some more details
	details, err := PlaceDetail(result.PlaceID)
	if err != nil {
		fmt.Printf("Could not fetch place detail for %d: %s\n", result.PlaceID, err)
		return
	}

	intro, extract, err := wikipedia.FetchIntro(details.Name)
	if err != nil {
		fmt.Printf("Could not fetch intro from Wikipedia for %s: %s\n", details.Name, err)
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
		attachment.Text = fmt.Sprintf("%s <%s/%s|_Wikipedia_>", intro, "https://de.wikipedia.org/wiki", extract.Title)
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

func replyWithError(something Something, err error) {
	fmt.Printf("An error occured: %s\n", err)

	itemRef := slack.NewRefToMessage(something.Channel, something.Msg.Timestamp)

	if err := api.AddReaction("question", itemRef); err != nil {
		fmt.Printf("An error occured while adding the reaction: %s\n", err)
	}
}
