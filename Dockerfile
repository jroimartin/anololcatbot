FROM golang:1.17-alpine3.14 as build

WORKDIR /go/src/github.com/jroimartin/anololcatbot
COPY . .

RUN go build ./cmd/anololcatbot


FROM alpine:3.14

RUN apk --update add ca-certificates

COPY --from=build \
	/go/src/github.com/jroimartin/anololcatbot/anololcatbot \
	/usr/local/bin/anololcatbot
RUN chmod 755 /usr/local/bin/anololcatbot

ENTRYPOINT ["/usr/local/bin/anololcatbot"]
