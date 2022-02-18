package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/slack-go/slack"
)

const emojiKeepInterval = time.Hour * 24

// EmojiManager manages emoji list
type EmojiManager struct {
	Client        *slack.Client
	EmojiList     []string
	LastUpdatedAt time.Time
}

// GetRandomEmoji is the function to get random emoji which is in the periodically updated EmojiList
func (em *EmojiManager) GetRandomEmoji() (string, error) {
	if em.LastUpdatedAt.IsZero() || time.Now().After(em.LastUpdatedAt.Add(emojiKeepInterval)) {
		em.UpdateEmojiList()
	}

	if em.EmojiList == nil || len(em.EmojiList) <= 0 {
		return "", errors.New("EmojiList Error")
	}

	return em.EmojiList[rand.Intn(len(em.EmojiList))], nil
}

// UpdateEmojiList is the function to renew emoji list
func (em *EmojiManager) UpdateEmojiList() error {
	emojiList, err := GetEmojiList(em.Client)
	if err != nil {
		return err
	}
	em.EmojiList = emojiList
	em.LastUpdatedAt = time.Now()
	return nil
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
