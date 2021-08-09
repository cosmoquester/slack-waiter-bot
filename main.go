package main

import (
	"fmt"
	"net/http"
	"os"
	"slack-waiter-bot/service"

	"math/rand"
	"time"
)

func main() {
	// "Bot User OAuth Access Token" which starts with "xoxb-"
	slackBotToken := os.Getenv("SLACK_BOT_USER_TOKEN")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	client, botUserID, err := service.AuthorizeSlack(slackBotToken)
	if err != nil {
		_ = fmt.Errorf("INVALID TOKEN ERROR")
		return
	}
	emojiList, err := service.GetEmojiList(client)
	if err != nil {
		_ = fmt.Errorf("INVALID EMOTION PERMISSION")
	}

	rand.Seed(time.Now().Unix())

	eventHandler := service.EventHandler{
		Client:        client,
		SigningSecret: signingSecret,
		BotUserID:     botUserID,
	}
	actionHandler := service.ActionHandler{
		Client:    client,
		EmojiList: emojiList,
	}

	http.HandleFunc("/events", eventHandler.HandleEvent)
	http.HandleFunc("/actions", actionHandler.HandleAction)

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
