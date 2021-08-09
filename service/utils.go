package service

import "github.com/slack-go/slack"

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

// GetEmojiList is the function to get emoji list from workspace
func GetEmojiList(client *slack.Client) ([]string, error) {
	emojiMap, err := client.GetEmoji()
	if err != nil {
		return nil, err
	}

	emojiList := make([]string, len(emojiMap))
	i := 0
	for k := range emojiMap {
		emojiList[i] = ":" + k + ": "
		i++
	}
	return emojiList, nil
}
