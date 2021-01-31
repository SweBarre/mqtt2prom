FROM golang:1.15.7-alpine3.13
RUN apk add --no-cache git
ADD . /go/src/mqtt2prom
WORKDIR /go/src/mqtt2prom
RUN go get mqtt2prom
RUN go install
WORKDIR /root
CMD ["/go/bin/mqtt2prom", "--config", "/etc/mqtt2prom.yaml"]
