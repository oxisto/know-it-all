package rest

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"encoding/json"
)

func SlackCallback(w http.ResponseWriter, r *http.Request) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	payload := r.FormValue("payload")

	fmt.Println(payload)

	trigger := map[string]interface{}{}

	err = json.Unmarshal([]byte(payload), &trigger)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(trigger)
}

