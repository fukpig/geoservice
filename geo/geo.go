package geoservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/go-redis/redis"
)

type tripDistanceInfo struct {
	Service  string
	Duration int32
	Distance int32
	Err      error
}

type externalGeoService interface {
	getTripInfo(from, to string) *tripDistanceInfo
}

func getRouteFromRedis(redisClient *redis.Client, from, to string) *tripDistanceInfo {
	val, err := redisClient.Get(from + ":" + to).Result()
	if err != nil {
		return &tripDistanceInfo{Service: "", Duration: 0, Distance: 0, Err: errors.New("No redis route")}
	}
	result := &tripDistanceInfo{}
	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return &tripDistanceInfo{Service: "", Duration: 0, Distance: 0, Err: errors.New("No redis route")}
	}
	fmt.Println(from+":"+to, val)
	return result
}

func setRouteFromRedis(redisClient *redis.Client, tripDistanceInfo *tripDistanceInfo, from, to string) error {
	tripInfoJson, err := json.Marshal(tripDistanceInfo)
	if err != nil {
		return err
	}

	err = redisClient.Set(from+":"+to, string(tripInfoJson), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func getFastestAnswer(from, to string) chan *tripDistanceInfo {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var mutex = &sync.Mutex{}

	distanceInfoChan := make(chan *tripDistanceInfo, 1)
	externalApis := [2]externalGeoService{googleGeoService{}, openstreetmapGeoService{}}

	for _, api := range externalApis {
		go func(ctx context.Context, infoChan chan *tripDistanceInfo, from, to string) {
			select {
			case <-ctx.Done():
				return
			default:
				var geoservice externalGeoService
				geoservice = api
				mutex.Lock()
				distanceInfoChan <- geoservice.getTripInfo(from, to)
				mutex.Unlock()
				cancel()
				//close(distanceInfoChan)
				return
			}
		}(ctx, distanceInfoChan, from, to)
	}

	return distanceInfoChan
}

func fallback(serviceName, from, to string) *tripDistanceInfo {
	var geoservice externalGeoService
	var fallbackResult *tripDistanceInfo
	if serviceName == "Google" {
		geoservice = openstreetmapGeoService{}
		fallbackResult = geoservice.getTripInfo(from, to)
	} else {
		geoservice = googleGeoService{}
		fallbackResult = geoservice.getTripInfo(from, to)
	}
	return fallbackResult
}

func Execute(redisClient *redis.Client, from, to string) *tripDistanceInfo {
	var result *tripDistanceInfo
	distanceInfoChan := getFastestAnswer(from, to)

	redisInfo := getRouteFromRedis(redisClient, from, to)
	if redisInfo.Err == nil {
		return redisInfo
	}

	info := <-distanceInfoChan

	if info.Err != nil {
		result = fallback(info.Service, from, to)
	} else {
		result = info
	}

	if result.Err != nil {
		return result
	}

	setRouteFromRedis(redisClient, result, from, to)

	return result

}
