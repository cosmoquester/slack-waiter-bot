package service

import (
	"encoding/json"
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
			messages, _, _, _ := eh.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: ev.Channel, Timestamp: ev.TimeStamp})
			for _, msg := range messages {
				if msg.User == eh.BotUserID {
					return
				}
			}

			headerText := slack.NewTextBlockObject("plain_text", "메뉴판", false, false)
			headerBlock := slack.NewHeaderBlock(headerText)

			addMenuBtnTxt := slack.NewTextBlockObject("plain_text", "메뉴 추가", false, false)
			addMenuBtn := slack.NewButtonBlockElement(ids.AddMenuButton, ids.AddMenuButton, addMenuBtnTxt)
			terminateBtnTxt := slack.NewTextBlockObject("plain_text", "종료", false, false)
			terminateBtn := slack.NewButtonBlockElement(ids.TerminateMenu, ids.TerminateMenu, terminateBtnTxt).WithStyle(slack.StyleDanger)

			ButtonBlock := slack.NewActionBlock(ids.MenuButtonsBlock, addMenuBtn, terminateBtn)
			eh.Client.PostMessage(ev.Channel, slack.MsgOptionBlocks(headerBlock, slack.NewDividerBlock(), ButtonBlock), slack.MsgOptionTS(ev.TimeStamp))
		}
	}

}
