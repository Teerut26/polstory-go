package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"googlemaps.github.io/maps"
)

func GetLocation(latitude, longitude float64) (string, error) {
	if latitude == 0 || longitude == 0 {
		return "", nil
	}
	// create a new context
	ctx := context.Background()

	mapClient, mapErr := maps.NewClient(maps.WithAPIKey(os.Getenv("GOOGLE_MAPS_API_KEY")))
	if mapErr != nil {
		return "", mapErr
	}

	geocodeRequest := &maps.GeocodingRequest{
		LatLng: &maps.LatLng{
			Lat: latitude,
			Lng: longitude,
		},
	}
	geocodeResponse, geocodeErr := mapClient.ReverseGeocode(ctx, geocodeRequest)
	if geocodeErr != nil {
		return "", geocodeErr
	}

	city := string(geocodeResponse[len(geocodeResponse)-5].AddressComponents[0].LongName)
	subregion := strings.Replace(string(geocodeResponse[len(geocodeResponse)-3].AddressComponents[0].LongName), " District", "", -1)
	region := string(geocodeResponse[len(geocodeResponse)-2].AddressComponents[0].LongName)
	country := string(geocodeResponse[len(geocodeResponse)-1].AddressComponents[0].LongName)

	var locationFormat string
	if country == "Thailand" {
		if subregion == region {
			locationFormat = fmt.Sprintf("  %s, %s", city, subregion)
		} else {
			locationFormat = fmt.Sprintf("  %s, %s", subregion, region)
		}
	} else {
		locationFormat = fmt.Sprintf("  %s, %s", region, country)
	}

	if subregion == "" || region == "" {
		locationFormat = ""
	}
	return locationFormat, nil
}
