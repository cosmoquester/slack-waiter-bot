package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func showAddMenuModal(channelID, originalTs string) slack.ModalViewRequest {
	menuNameText := slack.NewTextBlockObject("plain_text", "MENU", false, false)
	menuNamePlaceholder := slack.NewTextBlockObject("plain_text", "ex) 회전초밥 32pc", false, false)
	menuNameElement := slack.NewPlainTextInputBlockElement(menuNamePlaceholder, "menu_submit")
	menuName := slack.NewInputBlock("menu_submit_block", menuNameText, menuNameElement)

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject("plain_text", "메뉴를 골라주세옹!", false, false)
	modalRequest.Close = slack.NewTextBlockObject("plain_text", "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject("plain_text", "Submit", false, false)
	modalRequest.CallbackID = fmt.Sprintf("%s\t%s", channelID, originalTs)
	modalRequest.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			menuName,
		},
	}

	return modalRequest
}

func main() {
	// "Bot User OAuth Access Token" which starts with "xoxb-"
	slackBotToken := os.Getenv("SLACK_BOT_USER_TOKEN")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	api := slack.New(slackBotToken)

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
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
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}
		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent
			switch ev := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				headerText := slack.NewTextBlockObject("plain_text", "메뉴판", false, false)
				headerBlock := slack.NewHeaderBlock(headerText)

				addMenuBtnTxt := slack.NewTextBlockObject("plain_text", "메뉴 추가", false, false)
				addMenuBtn := slack.NewButtonBlockElement("menu_add", "menu_add", addMenuBtnTxt)
				terminateBtnTxt := slack.NewTextBlockObject("plain_text", "종료", false, false)
				terminateBtn := slack.NewButtonBlockElement("terminate", "terminate", terminateBtnTxt)

				ButtonBlock := slack.NewActionBlock("menu_buttons", addMenuBtn, terminateBtn)
				api.PostMessage(ev.Channel, slack.MsgOptionBlocks(headerBlock, slack.NewDividerBlock(), ButtonBlock), slack.MsgOptionTS(ev.TimeStamp))
			}
		}

	})

	http.HandleFunc("/actions", func(w http.ResponseWriter, r *http.Request) {
		var payload slack.InteractionCallback
		err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
		if err != nil {
			fmt.Printf("Could not parse action response JSON: %v\n", err)
		}

		switch payload.Type {
		case slack.InteractionTypeBlockActions:
			if len(payload.ActionCallback.BlockActions) != 1 || payload.ActionCallback.BlockActions[0].BlockID != "menu_buttons" {
				break
			}
			switch payload.ActionCallback.BlockActions[0].ActionID {
			case "menu_add":
				api.OpenView(payload.TriggerID, showAddMenuModal(payload.Channel.ID, payload.Message.Timestamp))
			case "terminate":
				terminateText := slack.NewTextBlockObject("plain_text", "END Order", false, false)
				terminateBlock := slack.NewSectionBlock(terminateText, nil, nil)
				api.UpdateMessage(payload.Channel.ID, payload.Message.Timestamp, slack.MsgOptionBlocks(terminateBlock))
			}

		case slack.InteractionTypeViewSubmission:
			callbackInfo := strings.Split(payload.View.CallbackID, "\t")
			channel := callbackInfo[0]
			originalPostTimeStamp := callbackInfo[1]
			menuString := payload.View.State.Values["menu_submit_block"]["menu_submit"].Value

			menuText := slack.NewTextBlockObject("plain_text", menuString, false, false)
			menuBlock := slack.NewSectionBlock(menuText, nil, nil)
			api.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(menuBlock))
		}

	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
