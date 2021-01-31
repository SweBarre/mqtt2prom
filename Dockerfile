FROM golang:1.15.7-alpine3.13
RUN apk add --no-cache git
ADD src/ /go/src/mqtt2prom
WORKDIR /go/src/mqtt2prom
RUN go get mqtt2prom
RUN go install
WORKDIR /root
COPY entrypoint.sh /usr/local/bin
COPY mqtt2prom.yaml /etc/mqtt2prom.yaml
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["run"]
