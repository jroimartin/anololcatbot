# AnoLolcatBot

Telegram bot for [ANO](http://ano.lolcathost.org/).

## Run locally

### Native

```
$ go get github.com/jroimartin/anololcatbot/cmd/anololcatbot
$ ANOLOLCATBOT_TOKEN=<token> anololcatbot
```

### Docker

```
$ _script/build
$ docker run --rm -e ANOLOLCATBOT_TOKEN=<token> anololcatbot
```

## Run in Digital Ocean

Create docker-machine if needed:

```
$ docker-machine create --driver digitalocean --digitalocean-access-token <token> --digitalocean-region ams3 --digitalocean-size 512mb docker
```

Run anololcatbot:

```
$ eval $(docker-machine env docker)
$ _script/build
$ ANOLOLCATBOT_TOKEN=<token> _script/deploy
```

References:

* [Digital Ocean example](https://docs.docker.com/machine/examples/ocean/)
* [Driver options](https://docs.docker.com/machine/drivers/digital-ocean/#options)
