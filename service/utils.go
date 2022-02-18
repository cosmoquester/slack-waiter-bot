package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// URLVerification when setting event URI
func URLVerification(w http.ResponseWriter, body []byte) {
	var r *slackevents.ChallengeResponse
	err := json.Unmarshal([]byte(body), &r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text")
	w.Write([]byte(r.Challenge))
}

// AuthorizeSlack is the function to authorize with bot token and return client, bot user id
func AuthorizeSlack(slackBotToken string) (*slack.Client, string, error) {
	client := slack.New(slackBotToken)

	bot, err := client.AuthTest()
	if err != nil {
		return nil, "", err
	}
	botUserID := bot.UserID
	return client, botUserID, nil
}

// GetMessageFromTimeStamp retrieve slack message using channel and timestamp
func GetMessageFromTimeStamp(client *slack.Client, channelID string, timestamp string) *slack.Message {
	messages, _, _, _ := client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: channelID, Timestamp: timestamp})
	for _, message := range messages {
		if message.Timestamp == timestamp {
			return &message
		}
	}
	return nil
}

// WriteCallbackMetadata returns private metadata for add menu view
func WriteCallbackMetadata(channelID string, timestamp string) string {
	return fmt.Sprintf("%s\t%s", channelID, timestamp)
}

// ParseCallbackMetadata returns parsed informations of add menu view
func ParseCallbackMetadata(privateMetadata string) (string, string) {
	callbackInfo := strings.Split(privateMetadata, "\t")
	channel := callbackInfo[0]
	originalPostTimeStamp := callbackInfo[1]
	return channel, originalPostTimeStamp
}
