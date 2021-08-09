package service

import (
	"fmt"
	"math/rand"
	"slack-waiter-bot/ids"
	"sync"

	"github.com/slack-go/slack"
)

var messageUpdateMutex = &sync.Mutex{}

// AddMenu handles when user clicks addmenu button
func AddMenu(ah ActionHandler, payload slack.InteractionCallback) {
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

	ah.Client.OpenView(payload.TriggerID, modalRequest)
}

// TerminateMenu handles when user clicks terminate button
func TerminateMenu(ah ActionHandler, payload slack.InteractionCallback) {
	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()
	messages, _, _, _ := ah.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: payload.Channel.ID, Timestamp: payload.Message.Timestamp})
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
		ah.Client.UpdateMessage(payload.Channel.ID, msg.Timestamp, slack.MsgOptionBlocks(blocks...))
	}
}

// SelectMenuByUser handles when user select a menu
func SelectMenuByUser(ah ActionHandler, payload slack.InteractionCallback, selectedMenuName string) {
	ToggleMenuOfUser(ah.Client, payload.User.ID, payload.Channel.ID, payload.Message.Timestamp, selectedMenuName)
}

// SubmitMenuAdd handles when user submit menu add view
func SubmitMenuAdd(ah ActionHandler, payload slack.InteractionCallback) {
	channel, originalPostTimeStamp := ParseAddMenuMetadata(payload.View.PrivateMetadata)

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()
	messages, _, _, _ := ah.Client.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: channel, Timestamp: originalPostTimeStamp})

	for _, msg := range messages {
		if msg.Timestamp != originalPostTimeStamp {
			continue
		}

		menuString := payload.View.State.Values[ids.SubmitMenuInputBlock][ids.SubmitMenuInput].Value
		menuText := slack.NewTextBlockObject("plain_text", ah.EmojiList[rand.Intn(len(ah.EmojiList))]+menuString, true, false)
		selectText := slack.NewTextBlockObject("plain_text", "Select", false, false)
		menuUserSelectBlock := slack.NewSectionBlock(menuText, nil, slack.NewAccessory(slack.NewButtonBlockElement(ids.SelectMenuByUser, menuString, selectText)))
		menuSelectContextBlock := slack.NewContextBlock(ids.MenuSelectContextBlock+menuString, slack.NewTextBlockObject("plain_text", "0 Selected", false, false))

		// Append New Menu Block
		blocks := make([]slack.Block, len(msg.Blocks.BlockSet)+2)
		copy(blocks, msg.Blocks.BlockSet[:len(msg.Blocks.BlockSet)-2])
		blocks[len(blocks)-4] = menuUserSelectBlock
		blocks[len(blocks)-3] = menuSelectContextBlock
		copy(blocks[len(blocks)-2:], msg.Blocks.BlockSet[len(msg.Blocks.BlockSet)-2:])

		ah.Client.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(blocks...))

		// Select default selected users
		selectedUsers := payload.View.State.Values[ids.SubmitMenuSelectPeopleBlock][ids.SubmitMenuPeople].SelectedUsers
		for _, user := range selectedUsers {
			go ToggleMenuOfUser(ah.Client, user, channel, originalPostTimeStamp, menuString)
		}
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

		var block *slack.ContextBlock
		var blockID string
		var blockIndex int
		for i, curBlock := range msg.Blocks.BlockSet {
			if curBlock.BlockType() == slack.MBTContext {
				contextBlock := curBlock.(*slack.ContextBlock)
				if contextBlock.BlockID[len(ids.MenuSelectContextBlock):] == selectedMenuName {
					block = curBlock.(*slack.ContextBlock)
					blockID = contextBlock.BlockID
					blockIndex = i
				}
			}
		}

		elements := block.ContextElements.Elements
		isExist := false
		for i, curElement := range elements[:len(elements)-1] {
			switch element := curElement.(type) {
			case *slack.ImageBlockElement:
				if element.AltText == profile.RealName {
					elements = append(elements[:i], elements[i+1:]...)
					isExist = true
					break
				}
			}
		}
		if !isExist {
			elements = make([]slack.MixedElement, len(elements)+1)
			copy(elements, block.ContextElements.Elements)
			elements[len(elements)-2] = slack.NewImageBlockElement(profile.Image32, profile.RealName)
		}

		elements[len(elements)-1] = slack.NewTextBlockObject("plain_text", fmt.Sprintf("%d Selected", len(elements)-1), false, false)
		msg.Blocks.BlockSet[blockIndex] = slack.NewContextBlock(blockID, elements...)
		client.UpdateMessage(channelID, msg.Timestamp, slack.MsgOptionBlocks(msg.Blocks.BlockSet...))
	}
}
