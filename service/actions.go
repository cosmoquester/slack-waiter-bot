package service

import (
	"math/rand"
	"slack-waiter-bot/ids"
	"sync"

	"github.com/slack-go/slack"
)

var messageUpdateMutex = &sync.Mutex{}

// AddMenu handles when user clicks addmenu button
func AddMenu(handler Handler, payload slack.InteractionCallback) {
	// Menu Input Block
	menuNameText := slack.NewTextBlockObject("plain_text", "MENU", false, false)
	menuNamePlaceholder := slack.NewTextBlockObject("plain_text", "ex) 회전초밥 32pc", false, false)
	menuNameElement := slack.NewPlainTextInputBlockElement(menuNamePlaceholder, ids.SubmitMenuInput)
	menuName := slack.NewInputBlock(ids.SubmitMenuInputBlock, menuNameText, menuNameElement)

	// User Select Block
	userSelectText := slack.NewTextBlockObject("plain_text", "Selecting People", false, false)
	multiUserSelect := slack.NewOptionsMultiSelectBlockElement("multi_users_select", nil, ids.SubmitMenuPeople)
	multiUserSelect.InitialUsers = []string{payload.User.ID}
	userSelect := slack.NewInputBlock(ids.SubmitMenuSelectPeopleBlock, userSelectText, multiUserSelect)

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject("plain_text", "메뉴를 골라주세옹!", false, false)
	modalRequest.Close = slack.NewTextBlockObject("plain_text", "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject("plain_text", "Submit", false, false)
	modalRequest.CallbackID = ids.SubmitMenuCallback
	modalRequest.PrivateMetadata = WriteAddMenuMetadata(payload.Channel.ID, payload.Message.Timestamp)
	modalRequest.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			menuName, userSelect,
		},
	}

	handler.Client.OpenView(payload.TriggerID, modalRequest)
}

// TerminateMenu handles when user clicks terminate button
func TerminateMenu(handler Handler, payload slack.InteractionCallback) {
	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()
	messages, _, _, _ := handler.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: payload.Channel.ID, Timestamp: payload.Message.Timestamp})
	for _, msg := range messages {
		if msg.Timestamp != payload.Message.Timestamp || msg.ParentUserId != payload.User.ID {
			continue
		}

		blocks := []slack.Block{}
		for _, curBlock := range msg.Blocks.BlockSet {
			switch curBlock.BlockType() {
			case slack.MBTAction:
			case slack.MBTSection:
				curBlock.(*slack.SectionBlock).Accessory = nil
				blocks = append(blocks, curBlock)
			default:
				blocks = append(blocks, curBlock)
			}
		}
		handler.Client.UpdateMessage(payload.Channel.ID, msg.Timestamp, slack.MsgOptionBlocks(blocks...))
	}
}

// SelectMenuByUser handles when user select a menu
func SelectMenuByUser(handler Handler, payload slack.InteractionCallback, selectedMenuName string) {
	ToggleMenuOfUser(handler.Client, payload.User.ID, payload.Channel.ID, payload.Message.Timestamp, selectedMenuName)
}

// SubmitMenuAdd handles when user submit menu add view
func SubmitMenuAdd(handler Handler, payload slack.InteractionCallback) {
	channel, originalPostTimeStamp := ParseAddMenuMetadata(payload.View.PrivateMetadata)

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()
	messages, _, _, _ := handler.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: channel, Timestamp: originalPostTimeStamp})

	for _, msg := range messages {
		if msg.Timestamp != originalPostTimeStamp {
			continue
		}

		menuString := payload.View.State.Values[ids.SubmitMenuInputBlock][ids.SubmitMenuInput].Value
		menuBoard := ParseMenuBlocks(msg.Blocks.BlockSet)
		menuBoard.AddMenu(menuString, handler.EmojiList[rand.Intn(len(handler.EmojiList))])

		// Select default selected users
		selectedUsers := payload.View.State.Values[ids.SubmitMenuSelectPeopleBlock][ids.SubmitMenuPeople].SelectedUsers
		for _, user := range selectedUsers {
			profile, _ := handler.Client.GetUserProfile(&slack.GetUserProfileParameters{UserID: user})
			menuBoard.ToggleMenuByUser(profile, menuString)
		}
		handler.Client.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
	}
}

// ToggleMenuOfUser select or de-select menu of the user
func ToggleMenuOfUser(client *slack.Client, userID string, channelID string, TimeStamp string, selectedMenuName string) {
	profile, _ := client.GetUserProfile(&slack.GetUserProfileParameters{UserID: userID})

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()
	messages, _, _, _ := client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: channelID, Timestamp: TimeStamp})
	for _, msg := range messages {
		if msg.Timestamp != TimeStamp {
			continue
		}

		menuBoard := ParseMenuBlocks(msg.Blocks.BlockSet)
		menuBoard.ToggleMenuByUser(profile, selectedMenuName)
		client.UpdateMessage(channelID, msg.Timestamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
	}
}
