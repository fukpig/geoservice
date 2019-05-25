build:
	protoc -I. --go_out=plugins=grpc:. \
	  proto/tripInfo/tripInfo.proto

	docker build -t geoservice .
run:
	docker run -p 40000:40000 -e JAEGER_HOST=localhost -e JAEGER_PORT=6831 -e GOOGLE_MAPS_API_KEY=AIzaSyDoVIMRxxJkfA8--jFqaUoBrjTirGwDEzg geoservice
