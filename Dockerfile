FROM golang

RUN mkdir -p /go/src/github.com/fukpig

ADD . /go/src/github.com/fukpig

RUN go get  -t -v ./...
RUN go get  github.com/canthefason/go-watcher
RUN go install github.com/canthefason/go-watcher/cmd/watcher

ENTRYPOINT  watcher -run github.com/fukpig/geoservice/cmd  -watch github.com/fukpig/geoservice
