package rest

import (
	"net/http"
	"fmt"
	"encoding/json"
	"github.com/nlopes/slack"
	"github.com/oxisto/know-it-all/bot"
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

func MorePhotos(w http.ResponseWriter, r *http.Request) {
	payload := r.FormValue("payload")

	callback := slack.AttachmentActionCallback{} //map[string]interface{}{}

	err := json.Unmarshal([]byte(payload), &callback)

	if err != nil {
		JsonResponse(w, r, nil, err)
		return
	}

	placeID := callback.CallbackID

	details, err := bot.PlaceDetail(placeID)
	if err != nil {
		fmt.Printf("Could not fetch place detail for %d: %s\n", placeID, err)
		JsonResponse(w, r, nil, err)
		return
	}

	msg := bot.PreparePhotoMessage(details)

	JsonResponse(w, r, msg, nil)
}
