package main

import (
	"log"
	"net/http"
	"os"
	"slack-waiter-bot/service"

	"math/rand"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// "Bot User OAuth Access Token" which starts with "xoxb-"
	slackBotToken := os.Getenv("SLACK_BOT_USER_TOKEN")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	client, botUserID, err := service.AuthorizeSlack(slackBotToken)
	if err != nil {
		logger.Fatal("[FATAL] INVALID TOKEN ERROR")
	}
	emojiList, err := service.GetEmojiList(client)
	if err != nil {
		logger.Fatal("[FATAL] INVALID EMOTION PERMISSION")
	}

	rand.Seed(time.Now().Unix())

	handler := &service.Handler{
		Client:        client,
		SigningSecret: signingSecret,
		BotUserID:     botUserID,
		EmojiList:     emojiList,
		Logger:        logger,
	}

	http.HandleFunc("/events", handler.HandleEvent)
	http.HandleFunc("/actions", handler.HandleAction)

	logger.Println("[INFO] Server listening")
	http.ListenAndServe(":8080", nil)
}
