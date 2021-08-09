package service

import (
	"slack-waiter-bot/ids"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// HandleAppMentionEvent handles when user mention bot
func HandleAppMentionEvent(ev *slackevents.AppMentionEvent, eh EventHandler) {
	var timeStamp string
	if ev.ThreadTimeStamp != "" {
		timeStamp = ev.ThreadTimeStamp
	} else {
		timeStamp = ev.TimeStamp
	}

	messages, _, _, _ := eh.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: ev.Channel, Timestamp: timeStamp})
	for _, msg := range messages {
		if msg.User == eh.BotUserID {
			return
		}
	}

	headerText := slack.NewTextBlockObject("plain_text", "메뉴판", false, false)
	headerBlock := slack.NewHeaderBlock(headerText)

	addMenuBtnTxt := slack.NewTextBlockObject("plain_text", "메뉴 추가", false, false)
	addMenuBtn := slack.NewButtonBlockElement(ids.AddMenu, ids.AddMenu, addMenuBtnTxt)
	terminateBtnTxt := slack.NewTextBlockObject("plain_text", "종료", false, false)
	terminateBtn := slack.NewButtonBlockElement(ids.TerminateMenu, ids.TerminateMenu, terminateBtnTxt).WithStyle(slack.StyleDanger)

	ButtonBlock := slack.NewActionBlock(ids.MenuButtonsBlock, addMenuBtn, terminateBtn)

	eh.Client.PostMessage(ev.Channel, slack.MsgOptionBlocks(headerBlock, slack.NewDividerBlock(), ButtonBlock), slack.MsgOptionTS(timeStamp))
}
