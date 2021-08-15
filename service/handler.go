package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	Logger        *log.Logger
}

// HandleEvent is the function to handle events
func (handler *Handler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handler.Logger.Println("[INFO] Bad request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sv, err := slack.NewSecretsVerifier(r.Header, handler.SigningSecret)
	if err != nil {
		handler.Logger.Println("[INFO] Bad request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := sv.Write(body); err != nil {
		handler.Logger.Println("[INFO] Bad request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := sv.Ensure(); err != nil {
		handler.Logger.Println("[INFO] Bad request")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		handler.Logger.Println("[INFO] Bad request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		handler.Logger.Println("[INFO] URL Verification")
		URLVerification(w, body)
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch event := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			handler.Logger.Println("[INFO] App mentioned event")
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
				handler.Logger.Println("[INFO] Add menu action")
				go AddMenu(handler, &payload)
			case ids.DeleteMenu:
				handler.Logger.Println("[INFO] Delete menu action")
				go DeleteMenu(handler, &payload)
			case ids.OrderForOther:
				handler.Logger.Println("[INFO] Order for other action")
				go OrderForOther(handler, &payload)
			case ids.TerminateMenu:
				handler.Logger.Println("[INFO] Terminate menu action")
				go TerminateMenu(handler, &payload)
			case ids.SelectMenuByUser:
				handler.Logger.Println("[INFO] Select menu action")
				go SelectMenuByUser(handler, &payload, blockAction.Value)
			}
		}
	case slack.InteractionTypeViewSubmission:
		switch payload.View.CallbackID {
		case ids.SubmitMenuCallback:
			handler.Logger.Println("[INFO] Submit menu add view")
			go SubmitMenuAdd(handler, &payload)
		case ids.SubmitOrderForOtherCallback:
			handler.Logger.Println("[INFO] Submit order for others view")
			go SubmitOrderForOther(handler, &payload)
		case ids.SubmitDeleteMenuCallback:
			handler.Logger.Println("[INFO] Submit delete menu view")
			go SubmitMenuDelete(handler, &payload)
		}

	}

}
