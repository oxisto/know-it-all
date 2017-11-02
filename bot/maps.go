package bot

import (
	"fmt"

	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

var c *maps.Client

func InitGoogleAPI(apiKey string) {
	var err error

	c, err = maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err)
		return
	}
}

func Geocode(address string) ([]maps.GeocodingResult, error) {
	r := &maps.GeocodingRequest{
		Address:  address,
		Language: "de",
	}

	return c.Geocode(context.Background(), r)
}

func PlaceDetail(placeID string) (maps.PlaceDetailsResult, error) {
	r := &maps.PlaceDetailsRequest{
		PlaceID:  placeID,
		Language: "de",
	}
	return c.PlaceDetails(context.Background(), r)
}
