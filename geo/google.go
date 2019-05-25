package geoservice

import (
	"context"
	"errors"
	"os"
	"time"

	"googlemaps.github.io/maps"
)

type googleGeoService struct {
}

func (s googleGeoService) getClient() (client *maps.Client, err error) {
	c, err := maps.NewClient(maps.WithAPIKey(os.Getenv("GOOGLE_MAPS_API_KEY")))
	return c, err
}

func (s googleGeoService) getTripInfo(from, to string) (result *tripDistanceInfo) {
	client, err := s.getClient()
	if err != nil {
		return &tripDistanceInfo{Service: "Google", Duration: 0, Distance: 0, Err: err}
	}

	r := &maps.DirectionsRequest{
		Origin:      from,
		Destination: to,
	}
	route, _, err := client.Directions(context.Background(), r)
	if err != nil {
		return &tripDistanceInfo{Service: "Google", Duration: 0, Distance: 0, Err: errors.New("No routes found")}
	}

	if len(route) == 0 {
		return &tripDistanceInfo{Service: "Google", Duration: 0, Distance: 0, Err: errors.New("No routes found")}
	}

	if len(route[0].Legs) == 0 {
		return &tripDistanceInfo{Service: "Google", Duration: 0, Distance: 0, Err: errors.New("No routes found")}
	}

	duration := int32(route[0].Legs[0].Duration / time.Minute)
	distance := int32(route[0].Legs[0].Distance.Meters / 1000)

	result = &tripDistanceInfo{Service: "Google", Duration: duration, Distance: distance, Err: nil}

	return result
}
