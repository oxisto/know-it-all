package rest

import (
	"net/http"
	"net/http/httputil"
	"fmt"
	"encoding/json"
	"github.com/nlopes/slack"
)

func JsonResponse(w http.ResponseWriter, r *http.Request, object interface{}, err error) {
	// uh-uh, we have an error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return not found if object is nil
	if object == nil {
		http.NotFound(w, r)
		return
	}

	// otherwise, lets try to decode the JSON
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(object); err != nil {
		// uh-uh we couldn't decode the JSON
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SlackCallback(w http.ResponseWriter, r *http.Request) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		JsonResponse(w, r, nil, err)
		return
	}
	fmt.Println(string(requestDump))

	payload := r.FormValue("payload")

	fmt.Println(payload)

	callback := slack.AttachmentActionCallback{} //map[string]interface{}{}

	err = json.Unmarshal([]byte(payload), &callback)

	if err != nil {
		JsonResponse(w, r, nil, err)
		return
	}

	// create a new message
	msg := slack.PostMessageParameters{}
	msg.Text = "test"

	JsonResponse(w, r, msg, nil)
}
