package google

import (
	"fmt"

	"googlemaps.github.io/maps"
	"net/url"
	"context"
	"strconv"
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
	// statically center around Munich for now
	bounds := maps.LatLngBounds{
		NorthEast: maps.LatLng{
			Lat: 48.2436429,
			Lng: 11.7890959,
		},
		SouthWest: maps.LatLng{
			Lat: 48.0638393,
			Lng: 11.3154977,
		},
	}

	r := &maps.GeocodingRequest{
		Address:  address,
		Region: "de",
		Language: "de",
		Bounds: &bounds,
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

func StaticMapUrl(location maps.LatLng, zoom int) string {
	center := fmt.Sprintf("%f,%f", location.Lat, location.Lng)

	query := url.Values{}
	query.Set("center", center)
	if zoom != -1 {
		query.Set("zoom", strconv.Itoa(zoom))
	}
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
