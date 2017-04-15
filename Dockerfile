FROM alpine:latest

RUN apk --update add ca-certificates

ADD _build/anololcatbot /opt/anololcatbot/anololcatbot
RUN chmod 755 /opt/anololcatbot/anololcatbot

ENTRYPOINT ["/opt/anololcatbot/anololcatbot"]
