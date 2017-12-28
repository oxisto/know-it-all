package bot

import (
	"fmt"

	"strings"

	"errors"

	"math/rand"

	"github.com/nlopes/slack"
	"github.com/oxisto/know-it-all/wikipedia"
	"googlemaps.github.io/maps"
	"github.com/oxisto/know-it-all/google"
	"log"
)

type Something struct {
	Channel string
	Msg     slack.Msg
}

type State struct {
	DirectMessagesOnly bool
	BotId              string
	ReplyChannel       chan Something
}

var state *State
var api *slack.Client

func InitBot(token string, directMessagesOnly bool) {
	state = &State{
		DirectMessagesOnly: directMessagesOnly,
	}

	log.Println("Connecting to Slack...")

	api = slack.New(token)

	api.SetDebug(true)

	state.ReplyChannel = make(chan Something)
	go handleBotReply()

	/*rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				state.BotId = ev.Info.User.ID
				log.Printf("Connected to Slack as %s\n", state.BotId)
			case *slack.MessageEvent:
				something := Something{
					Msg:     ev.Msg,
					Channel: ev.Channel,
				}

				if ev.Msg.User != state.BotId && (!state.DirectMessagesOnly || state.DirectMessagesOnly && strings.HasPrefix(ev.Channel, "D")) {
					log.Printf("Handling message from %s...\n", ev.Channel)
					state.ReplyChannel <- something
				}
			case *slack.ReactionAddedEvent:
				// Handle reaction added
			case *slack.ReactionRemovedEvent:
				// Handle reaction removed
			case *slack.RTMError:
				log.Printf("Error: %s\n", ev.Error())
			}
		}
	}*/
}

func handleBotReply() {
	for {
		something := <-state.ReplyChannel

		tokens := TokenizeAndNormalize(something.Msg.Text)

		if index := tokens.ContainsTuple([]string{"wo", "ist"}); index != -1 {
			locationCommand(something, tokens, index)
		} else if index := tokens.ContainsTuple([]string{"wo", "sind"}); index != -1 {
			locationCommand(something, tokens, index)
		} else if index := tokens.ContainsTuple([]string{"wo", "liegt"}); index != -1 {
			locationCommand(something, tokens, index)
		} else if index := tokens.ContainsTuple([]string{"wo", "liegen"}); index != -1 {
			locationCommand(something, tokens, index)
		} else if index := tokens.ContainsTuple([]string{"was", "ist"}); index != -1 {
			lookupCommand(something, tokens, index)
		} else if index := tokens.ContainsTuple([]string{"was", "sind"}); index != -1 {
			lookupCommand(something, tokens, index)
		}
	}
}

type Tokens []string
type Token string
type InvertedIndex map[string][]int

func TokenizeAndNormalize(message string) Tokens {
	marks := []string{",", ".", "!", ":", "?"}
	fillerWords := []string{"eigentlich", "bitte", "danke", "der", "die", "das", "ein", "denn"}

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

func (tokens Tokens) ContainsTuple(words []string) (index int) {
	index = -1
	for i, w := range tokens {
		if w == words[0] {
			if tokens[i+1] != "" && tokens[i+1] == words[1] {
				index = i
			}
		}
	}

	return index
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
	log.Println("Triggering lookup command...")

	thing := tokens[index+2:].Reassemble()

	if thing == "" {
		// just ignore it
		return
	}

	log.Printf("Trying to lookup %s...\n", thing)

	intro, extract, err := wikipedia.FetchIntro(thing)
	if err != nil {
		log.Printf("Could not fetch intro from Wikipedia for %s: %s\n", thing, err)
	}

	if intro == "" {
		replyWithError(something, errors.New(fmt.Sprintf("Could not find %s", thing)))
		return
	}

	attachment := slack.Attachment{
		Color:      "#B733FF",
		Title:      extract.Title,
		TitleLink:  fmt.Sprintf("https://de.wikipedia.org/wiki/%s", extract.Title),
		Text:       fmt.Sprintf("%s <%s/%s|_Wikipedia_>", intro, "https://de.wikipedia.org/wiki", extract.Title),
		MarkdownIn: []string{"text"},
	}

	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.Attachments = []slack.Attachment{attachment}

	api.PostMessage(something.Channel, "", params)
}

func locationCommand(something Something, tokens Tokens, index int) {
	log.Println("Triggering location command...")

	locationName := tokens[index+2:].Reassemble()

	if locationName == "" {
		// just ignore it
		return
	}

	log.Printf("Trying to locate %s...\n", locationName)

	results, err := google.Geocode(locationName)
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
	details, err := google.PlaceDetail(result.PlaceID)
	if err != nil {
		log.Printf("Could not fetch place detail for %d: %s\n", result.PlaceID, err)
		return
	}

	attachment := slack.Attachment{
		Color:     "#B733FF",
		Title:     details.Name,
		TitleLink: details.URL,
	}

	// no zoom per default
	var zoom = -1
	if ContainsType(details.Types, "political") {
		// fetch intro from Wikipedia
		intro, extract, err := wikipedia.FetchIntro(details.Name)
		if err != nil {
			log.Printf("Could not fetch intro from Wikipedia for %s: %s\n", details.Name, err)
		}

		if intro != "" {
			attachment.Text = fmt.Sprintf("%s <%s/%s|_Wikipedia_>", intro, "https://de.wikipedia.org/wiki", extract.Title)
			attachment.MarkdownIn = []string{"text"}
		}

		// larger zoom for cities and countries
		zoom = 9
	} else {
		// otherwise, write address and reviews
		attachment.Text = details.FormattedAddress + "\n" + details.Website + "\n" + details.InternationalPhoneNumber

		// add rating
		attachment.Text += fmt.Sprintf("\n%.2f", details.Rating)

		// add stars
		for i := 0; i < int(details.Rating); i++ {
			attachment.Text += ":star:"
		}

		// add number of reviews
		attachment.Text += fmt.Sprintf(" %d Erfahrungsberichte", len(details.Reviews))
	}
	attachment.ImageURL = google.StaticMapUrl(result.Geometry.Location, zoom)

	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.Attachments = []slack.Attachment{attachment}

	api.PostMessage(something.Channel, "", params)

	if len(details.Photos) > 0 {
		api.PostMessage(something.Channel, "", PreparePhotoMessage(details))
	}
}

func replyWithError(something Something, err error) {
	log.Printf("An error occured: %s\n", err)

	itemRef := slack.NewRefToMessage(something.Channel, something.Msg.Timestamp)

	if err := api.AddReaction("question", itemRef); err != nil {
		log.Printf("An error occured while adding the reaction: %s\n", err)
	}
}

func SendMessage(channel string, text string, params slack.PostMessageParameters) {
	api.PostMessage(channel, text, params)
}

func PreparePhotoMessage(details maps.PlaceDetailsResult) slack.PostMessageParameters {
	actions := []slack.AttachmentAction{
		{
			Name: "more",
			Text: "Weitere Bilder",
			Type: "button",
		},
	}

	attachment := slack.Attachment{
		Color:      "#B733FF",
		Title:      fmt.Sprintf("Impressionen aus %s", details.Name),
		ImageURL:   google.PhotoUrl(details.Photos[rand.Int31n(int32(len(details.Photos)))].PhotoReference),
		CallbackID: details.PlaceID,
		Actions:    actions,
	}

	params := slack.PostMessageParameters{}
	params.AsUser = true
	params.Attachments = []slack.Attachment{attachment}

	return params
}

func ContainsType(types []string, t string) bool {
	for _, v := range types {
		if v == t {
			return true
		}
	}

	return false
}
