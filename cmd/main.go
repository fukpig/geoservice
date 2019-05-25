package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"contrib.go.opencensus.io/exporter/jaeger"
	geoservice "github.com/fukpig/geoservice/geo"
	pb "github.com/fukpig/geoservice/proto/tripInfo"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
)

type service struct {
	redisClient *redis.Client
}

func (s *service) GetTripInfo(ctx context.Context, req *pb.Route) (*pb.Response, error) {
	spanContextJson := req.SpanContext
	var spanContext trace.SpanContext
	err := json.Unmarshal([]byte(spanContextJson), &spanContext)
	if err != nil {
		logrus.Errorf("GetTripInfo, json.Unmarshal: %v\n", err)
		return &pb.Response{}, err
	}

	_, span := trace.StartSpanWithRemoteParent(context.Background(), "GeoService", spanContext)
	defer span.End()

	tripInfo := geoservice.Execute(s.redisClient, req.From, req.To)

	if tripInfo.Err != nil {
		return &pb.Response{}, tripInfo.Err
	} else {
		return &pb.Response{Distance: tripInfo.Distance, Duration: tripInfo.Duration}, nil
	}
}

func InitJaeger(serviceName string) error {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: os.Getenv("JAEGER_HOST") + ":" + os.Getenv("JAEGER_PORT"),
		Process: jaeger.Process{
			ServiceName: serviceName,
			Tags: []jaeger.Tag{
				jaeger.StringTag("hostname", "localhost"),
			},
		},
	})
	if err != nil {
		return err
	}
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "localhost:9003")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	InitJaeger("taxi-geo-service")

	s := grpc.NewServer()
	pb.RegisterGeoServiceServer(s, &service{redisClient: redisClient})

	log.Println("Running on port:", "9003")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
