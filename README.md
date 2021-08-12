# slack-waiter-bot

[![cosmoquester](https://circleci.com/gh/cosmoquester/slack-waiter-bot.svg?style=svg)](https://app.circleci.com/pipelines/github/cosmoquester/slack-waiter-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmoquester/slack-waiter-bot)](https://goreportcard.com/report/github.com/cosmoquester/slack-waiter-bot)
[![codecov](https://codecov.io/gh/cosmoquester/slack-waiter-bot/branch/master/graph/badge.svg?token=B8MCqXb1bZ)](https://codecov.io/gh/cosmoquester/slack-waiter-bot)

This is WaiterBot Server which is go study project in ScatterLab.

## Run

```sh
$ docker run \
    -e SLACK_SIGNING_SECRET=somtehin124singnssecret \
    -e SLACK_BOT_USER_TOKEN=xoxb-123124412-1231231231231-dsfapodfjasdi \
    cosmoquester/slack-waiter-bot
```

## Settings

### URL setting required on Slack Bot setting

- Interactivity & Shortcuts Request URL: http://[SERVER-URI]/actions
- Event Subscriptions Request URL: http://[SERVER-URI]/events

### Add permissions below to Bot Token Scopes

- channels:history
- chat:write
- chat:write.public
- conversations.connect:write
- emoji:read
- groups:history
- im:history
- im:write
- mpim:history
- users.profile:read
