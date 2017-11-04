package google

import (
	"fmt"

	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
	"net/url"
)

var c *maps.Client

var key string

func InitAPI(apiKey string) {
	var err error

	key = apiKey

	c, err = maps.NewClient(maps.WithAPIKey(key))
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

func StaticMapUrl(location maps.LatLng) string {
	center := fmt.Sprintf("%f,%f", location.Lat, location.Lng)

	query := url.Values{}
	query.Set("center", center)
	query.Set("zoom", "9")
	query.Set("markers", fmt.Sprintf("color:red|label:|%s", center))
	query.Set("size", "640x480")
	query.Set("key", key)

	return "https://maps.googleapis.com/maps/api/staticmap?" + query.Encode()
}

func PhotoUrl(photoReference string) string {
	query := url.Values{}
	query.Set("maxwidth", "400")
	query.Set("photoreference", photoReference)
	query.Set("key", key)

	return "https://maps.googleapis.com/maps/api/place/photo?" + query.Encode()
}
