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

// Action IDs
const (
	SelectMenuByUser = "select_menu_by_user"
	SubmitMenu       = "submit_menu"
	AddMenuButton    = "add_menu"
	TerminateMenu    = "terminate_menu"
)

// Block IDs
const (
	SubmitMenuBlock  = "submit_menu_block"
	MenuButtonsBlock = "menu_buttons_block"
)

func showAddMenuModal(channelID, originalTs string) slack.ModalViewRequest {
	menuNameText := slack.NewTextBlockObject("plain_text", "MENU", false, false)
	menuNamePlaceholder := slack.NewTextBlockObject("plain_text", "ex) 회전초밥 32pc", false, false)
	menuNameElement := slack.NewPlainTextInputBlockElement(menuNamePlaceholder, SubmitMenu)
	menuName := slack.NewInputBlock(SubmitMenuBlock, menuNameText, menuNameElement)

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
				addMenuBtn := slack.NewButtonBlockElement(AddMenuButton, AddMenuButton, addMenuBtnTxt)
				terminateBtnTxt := slack.NewTextBlockObject("plain_text", "종료", false, false)
				terminateBtn := slack.NewButtonBlockElement(TerminateMenu, TerminateMenu, terminateBtnTxt)

				ButtonBlock := slack.NewActionBlock(MenuButtonsBlock, addMenuBtn, terminateBtn)
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
			if len(payload.ActionCallback.BlockActions) != 1 || payload.ActionCallback.BlockActions[0].BlockID != MenuButtonsBlock {
				break
			}
			switch payload.ActionCallback.BlockActions[0].ActionID {
			case AddMenuButton:
				api.OpenView(payload.TriggerID, showAddMenuModal(payload.Channel.ID, payload.Message.Timestamp))
			case TerminateMenu:
				terminateText := slack.NewTextBlockObject("plain_text", "END Order", false, false)
				terminateBlock := slack.NewSectionBlock(terminateText, nil, nil)
				api.UpdateMessage(payload.Channel.ID, payload.Message.Timestamp, slack.MsgOptionBlocks(terminateBlock))
			}

		case slack.InteractionTypeViewSubmission:
			callbackInfo := strings.Split(payload.View.CallbackID, "\t")
			channel := callbackInfo[0]
			originalPostTimeStamp := callbackInfo[1]

			messages, _, _, _ := api.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: channel, Timestamp: originalPostTimeStamp})

			for _, msg := range messages {
				if msg.Timestamp != originalPostTimeStamp {
					continue
				}

				menuString := payload.View.State.Values[SubmitMenuBlock][SubmitMenu].Value
				menuText := slack.NewTextBlockObject("plain_text", menuString, false, false)
				selectText := slack.NewTextBlockObject("plain_text", "Select", false, false)
				menuUserSelectBlock := slack.NewSectionBlock(menuText, nil, slack.NewAccessory(slack.NewButtonBlockElement(SelectMenuByUser, menuString, selectText)))

				// Append New Menu Block
				blocks := make([]slack.Block, len(msg.Blocks.BlockSet)+1)
				copy(blocks, msg.Blocks.BlockSet[:len(msg.Blocks.BlockSet)-2])
				blocks[len(blocks)-3] = menuUserSelectBlock
				copy(blocks[len(blocks)-2:], msg.Blocks.BlockSet[len(msg.Blocks.BlockSet)-2:])

				api.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(blocks...))
			}

		}

	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
