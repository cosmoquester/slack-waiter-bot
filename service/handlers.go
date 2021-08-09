package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"slack-waiter-bot/ids"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// EventHandler for handling slack events
type EventHandler struct {
	Client        *slack.Client
	SigningSecret string
	BotUserID     string
}

// ActionHandler for handling slack actions
type ActionHandler struct {
	Client    *slack.Client
	EmojiList []string
}

// HandleEvent is the function to handle events
func (eh EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sv, err := slack.NewSecretsVerifier(r.Header, eh.SigningSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		URLVerification(w, body)
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			HandleAppMentionEvent(event, eh)
		}
	}

}

// HandleAction is the function to handle actions
func (ah ActionHandler) HandleAction(w http.ResponseWriter, r *http.Request) {
	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v\n", err)
	}

	switch payload.Type {
	case slack.InteractionTypeBlockActions:
		for _, blockAction := range payload.ActionCallback.BlockActions {
			switch blockAction.ActionID {
			case ids.AddMenu:
				AddMenu(ah, payload)
			case ids.TerminateMenu:
				TerminateMenu(ah, payload)
			case ids.SelectMenuByUser:
				SelectMenuByUser(ah, payload, blockAction.Value)
			}
		}
	case slack.InteractionTypeViewSubmission:
		switch payload.View.CallbackID {
		case ids.SubmitMenu:
			SubmitMenuAdd(ah, payload)
		}
	}

}
