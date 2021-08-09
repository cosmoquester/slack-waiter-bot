# slack-waiter-bot

[![cosmoquester](https://circleci.com/gh/cosmoquester/slack-waiter-bot.svg?style=svg)](https://app.circleci.com/pipelines/github/cosmoquester/slack-waiter-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/cosmoquester/slack-waiter-bot)](https://goreportcard.com/report/github.com/cosmoquester/slack-waiter-bot)
[![codecov](https://codecov.io/gh/cosmoquester/slack-waiter-bot/branch/master/graph/badge.svg?token=B8MCqXb1bZ)](https://codecov.io/gh/cosmoquester/slack-waiter-bot)

스캐터랩 사내 go 스터디 프로젝트로 진행하는 웨이터봇입니다.

## 실행

```sh
$ docker run \
    -e SLACK_SIGNING_SECRET=somtehin124singnssecret \
    -e SLACK_BOT_USER_TOKEN=xoxb-123124412-1231231231231-dsfapodfjasdi \
    cosmoquester/slack-waiter-bot
```

## 설정

Slack 봇 설정에 들어가서

- Interactivity & Shortcuts Request URL: http://[SERVER-DOMAIN]/actions
- Event Subscriptions Request URL: http://[SERVER-DOMAIN]/events

위와 같이 설정해주어야합니다.
