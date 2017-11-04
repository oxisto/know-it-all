package wikipedia

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

type QueryResponse struct {
	Query Query `json:"query"`
}

type ParseResponse struct {
	Parse Parse `json:"parse"`
}

type Parse struct {
	PageID   int               `json:"pageid"`
	Title    string            `json:"title"`
	WikiText map[string]string `json:"wikitext"`
}

type Query struct {
	Pages map[string]PageExtract `json:"pages"`
}

type PageExtract struct {
	PageID  int    `json:"pageid"`
	Title   string `json:"title"`
	Extract string `json:"extract"`
}

func FetchInfoBox(pageId int) (map[string]string, error) {
	u := fmt.Sprintf("https://de.wikipedia.org/w/api.php?action=parse&pageid=%d&section=0&prop=wikitext&format=json", pageId)

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	response := ParseResponse{}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	wikiText := strings.Replace(response.Parse.WikiText["*"], "\n", "", -1)

	r, _ := regexp.Compile("{{Infobox Gemeinde in Deutschland(.[^}]*)}}")
	m := r.FindStringSubmatch(wikiText)

	if len(m) > 0 {
		text := m[1]

		r, _ = regexp.Compile("(.[^= ]*) = (.[^|]*)")
		all := r.FindAllStringSubmatch(text, -1)

		infoBox := map[string]string{}

		for _, v := range all {
			key := v[1][1:]
			value := v[2]

			infoBox[key] = value
		}

		return infoBox, nil
	} else {
		return nil, nil
	}
}

func FetchIntro(page string) (string, *PageExtract, error) {
	u := fmt.Sprintf("https://de.wikipedia.org/w/api.php?format=json&action=query&prop=extracts&exintro=&explaintext=&titles=%s", url.PathEscape(page))

	res, err := http.Get(u)
	if err != nil {
		return "", nil, err
	}

	response := QueryResponse{}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return "", nil, err
	}

	keys := reflect.ValueOf(response.Query.Pages).MapKeys()
	extract := response.Query.Pages[keys[0].String()]

	intro := strings.Split(extract.Extract, "\n")[0]

	return intro, &extract, nil
}
