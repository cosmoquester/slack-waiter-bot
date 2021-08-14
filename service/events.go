package service

import (
	"slack-waiter-bot/ids"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// HandleAppMentionEvent handles when user mention bot
func HandleAppMentionEvent(event *slackevents.AppMentionEvent, eh *Handler) {
	var timeStamp string
	if event.ThreadTimeStamp != "" {
		timeStamp = event.ThreadTimeStamp
	} else {
		timeStamp = event.TimeStamp
	}

	messages, _, _, _ := eh.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: event.Channel, Timestamp: timeStamp})
	for _, msg := range messages {
		if msg.User == eh.BotUserID {
			return
		}
	}

	headerText := slack.NewTextBlockObject("plain_text", "Menu", false, false)
	headerBlock := slack.NewHeaderBlock(headerText)

	addMenuBtnTxt := slack.NewTextBlockObject("plain_text", "➕", false, false)
	addMenuBtn := slack.NewButtonBlockElement(ids.AddMenu, ids.AddMenu, addMenuBtnTxt)
	deleteMenuBtnTxt := slack.NewTextBlockObject("plain_text", "➖", false, false)
	deleteMenuBtn := slack.NewButtonBlockElement(ids.DeleteMenu, ids.DeleteMenu, deleteMenuBtnTxt)
	OrderForOtherBtnTxt := slack.NewTextBlockObject("plain_text", "👥", false, false)
	OrderForOtherBtn := slack.NewButtonBlockElement(ids.OrderForOther, ids.OrderForOther, OrderForOtherBtnTxt)
	terminateBtnTxt := slack.NewTextBlockObject("plain_text", "🚫", false, false)
	terminateBtn := slack.NewButtonBlockElement(ids.TerminateMenu, ids.TerminateMenu, terminateBtnTxt).WithStyle(slack.StyleDanger)

	ButtonBlock := slack.NewActionBlock(ids.MenuButtonsBlock, addMenuBtn, deleteMenuBtn, OrderForOtherBtn, terminateBtn)

	eh.Client.PostMessage(event.Channel, slack.MsgOptionBlocks(headerBlock, slack.NewDividerBlock(), ButtonBlock), slack.MsgOptionTS(timeStamp))
}
