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

// Handler for handling slack events and actions
type Handler struct {
	Client        *slack.Client
	SigningSecret string
	BotUserID     string
	EmojiList     []string
}

// HandleEvent is the function to handle events
func (handler *Handler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sv, err := slack.NewSecretsVerifier(r.Header, handler.SigningSecret)
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
			go HandleAppMentionEvent(event, handler)
		}
	}

}

// HandleAction is the function to handle actions
func (handler *Handler) HandleAction(w http.ResponseWriter, r *http.Request) {
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
				go AddMenu(handler, &payload)
			case ids.DeleteMenu:
				go DeleteMenu(handler, &payload)
			case ids.OrderForOther:
				go OrderForOther(handler, &payload)
			case ids.TerminateMenu:
				go TerminateMenu(handler, &payload)
			case ids.SelectMenuByUser:
				go SelectMenuByUser(handler, &payload, blockAction.Value)
			}
		}
	case slack.InteractionTypeViewSubmission:
		switch payload.View.CallbackID {
		case ids.SubmitMenuCallback:
			go SubmitMenuAdd(handler, &payload)
		case ids.SubmitOrderForOtherCallback:
			go SubmitOrderForOther(handler, &payload)
		case ids.SubmitDeleteMenuCallback:
			go SubmitMenuDelete(handler, &payload)
		}

	}

}
