FROM golang:1.15.7-alpine3.13
RUN apk add --no-cache git
ADD . /go/src/mqtt2prom
COPY entrypoint.sh /usr/local/bin
WORKDIR /go/src/mqtt2prom
RUN go get mqtt2prom
RUN go install
WORKDIR /root
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["run"]
