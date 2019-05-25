package geoservice

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

type point struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type routesInfo struct {
	Duration float64 `json:"duration"`
	Distance float64 `json:"distance"`
}

type routerAnswer struct {
	Routes []routesInfo `json:"routes"`
}

type openstreetmapGeoService struct {
	Client http.Client
}

func (s openstreetmapGeoService) getClient() {
	s.Client = http.Client{}
}

func (s openstreetmapGeoService) makeRequest(url string) (answer []byte, err error) {
	ctx := context.Background()
	ctxTimeout, timeoutCancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	defer timeoutCancel()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return make([]byte, 0), err
	}

	resp, err := s.Client.Do(req.WithContext(ctxTimeout))
	if err != nil {
		return make([]byte, 0), err
	}

	defer resp.Body.Close()
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return htmlBytes, nil
}

func (s openstreetmapGeoService) getCoordinatesFromAddress(address string) (p point, err error) {
	url := "https://nominatim.openstreetmap.org/search?q=" + address + "&format=json&polygon=1&addressdetails=1"
	answer, _ := s.makeRequest(url)
	points := make([]point, 0)
	err = json.Unmarshal(answer, &points)
	if len(points) == 0 {
		return point{}, errors.New("No location matches")
	}
	p = points[0]
	return p, err
}

func (s openstreetmapGeoService) getRouteInfo(fromPoint, toPoint point) (result *tripDistanceInfo) {
	url := "http://router.project-osrm.org/route/v1/driving/" + fromPoint.Lon + "," + fromPoint.Lat + ";" + toPoint.Lon + "," + toPoint.Lat + "?overview=false"
	answer, _ := s.makeRequest(url)
	var routerInfo routerAnswer
	err := json.Unmarshal(answer, &routerInfo)
	if err != nil {
		return &tripDistanceInfo{Service: "Openstreetmap", Duration: 0, Distance: 0, Err: err}
	}

	if len(routerInfo.Routes) == 0 {
		return &tripDistanceInfo{Service: "Openstreetmap", Duration: 0, Distance: 0, Err: errors.New("No routes found")}
	}

	duration := int32(math.Ceil(routerInfo.Routes[0].Duration / 60))
	distance := int32(math.Ceil(routerInfo.Routes[0].Distance / 1000))

	return &tripDistanceInfo{Service: "Openstreetmap", Duration: duration, Distance: distance, Err: err}
}

func (s openstreetmapGeoService) getTripInfo(from, to string) (result *tripDistanceInfo) {
	s.getClient()
	fromPoint, err := s.getCoordinatesFromAddress(from)
	if err != nil {
		return &tripDistanceInfo{Service: "Openstreetmap", Duration: 0, Distance: 0, Err: err}
	}

	toPoint, err := s.getCoordinatesFromAddress(to)
	if err != nil {
		return &tripDistanceInfo{Service: "Openstreetmap", Duration: 0, Distance: 0, Err: err}
	}

	tripDistanceInfo := s.getRouteInfo(fromPoint, toPoint)

	return tripDistanceInfo
}
