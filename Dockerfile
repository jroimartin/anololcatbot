FROM golang:1.14-alpine as build

RUN apk --update add git

WORKDIR /go/src/github.com/jroimartin/anololcatbot
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go get -v ./cmd/anololcatbot


FROM alpine:latest

RUN apk --update add ca-certificates

WORKDIR /opt/anololcatbot/
COPY --from=build /go/bin/anololcatbot anololcatbot
RUN chmod 755 anololcatbot

ENTRYPOINT ["/opt/anololcatbot/anololcatbot"]
