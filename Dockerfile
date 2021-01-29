FROM golang:alpine3.13

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o mqtt2prom
CMD ["/app/mqtt2prom"]
