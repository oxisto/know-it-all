package wikipedia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type ApiResponse struct {
	Query Query `json:"query"`
}

type Query struct {
	Pages map[string]PageExtract `json:"pages"`
}

type PageExtract struct {
	PageID  int    `json:"pageid"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

func FetchIntro(page string) (string, error) {
	u := fmt.Sprintf("https://de.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exintro=&explaintext=&titles=%s", url.PathEscape(page))

	res, err := http.Get(u)
	if err != nil {
		return "", err
	}

	response := ApiResponse{}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	keys := reflect.ValueOf(response.Query.Pages).MapKeys()

	extract := response.Query.Pages[keys[0].String()]

	intro := strings.Split(extract.Extract, "\n")[0]

	return intro, nil
}
