package service

import (
	"fmt"
	"math/rand"
	"slack-waiter-bot/ids"
	"strings"
	"sync"

	"github.com/slack-go/slack"
)

var messageUpdateMutex = &sync.Mutex{}

// AddMenu handles when user clicks addmenu button
func AddMenu(handler *Handler, payload *slack.InteractionCallback) {
	// Menu Input Block
	menuNameText := slack.NewTextBlockObject("plain_text", "메뉴를 골라달라옹", false, false)
	menuNamePlaceholder := slack.NewTextBlockObject("plain_text", "ex) 회전초밥 32pc", false, false)
	menuNameElement := slack.NewPlainTextInputBlockElement(menuNamePlaceholder, ids.SubmitMenuInput)
	menuName := slack.NewInputBlock(ids.SubmitMenuInputBlock, menuNameText, menuNameElement)

	// User Select Block
	userSelectText := slack.NewTextBlockObject("plain_text", "먹는 사람들도 골라달라옹", false, false)
	multiUserSelect := slack.NewOptionsMultiSelectBlockElement("multi_users_select", nil, ids.SubmitMenuPeople)
	multiUserSelect.InitialUsers = []string{payload.User.ID}
	userSelect := slack.NewInputBlock(ids.SubmitMenuSelectPeopleBlock, userSelectText, multiUserSelect)

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject("plain_text", "메뉴 추가", false, false)
	modalRequest.Close = slack.NewTextBlockObject("plain_text", "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject("plain_text", "Submit", false, false)
	modalRequest.CallbackID = ids.SubmitMenuCallback
	modalRequest.PrivateMetadata = WriteCallbackMetadata(payload.Channel.ID, payload.Message.Timestamp)
	modalRequest.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			menuName, userSelect,
		},
	}

	handler.Client.OpenView(payload.TriggerID, modalRequest)
}

// DeleteMenu handles when user clicks delete menu button
func DeleteMenu(handler *Handler, payload *slack.InteractionCallback) {
	message := GetMessageFromTimeStamp(handler.Client, payload.Channel.ID, payload.Message.Timestamp)
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)

	// Menu Input Block
	menuListText := slack.NewTextBlockObject("plain_text", "⚠️ 메뉴와 선택한 사람들이 모두 사라지니 조심해달라옹 ⚠️", false, false)
	menuListElement := slack.NewRadioButtonsBlockElement(ids.SubmitMenuInput, menuBoard.ToOptionBlockObjects()...)
	menuList := slack.NewInputBlock(ids.SubmitMenuDeleteBlock, menuListText, menuListElement)

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject("plain_text", "메뉴 삭제", false, false)
	modalRequest.Close = slack.NewTextBlockObject("plain_text", "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject("plain_text", "Submit", false, false)
	modalRequest.CallbackID = ids.SubmitDeleteMenuCallback
	modalRequest.PrivateMetadata = WriteCallbackMetadata(payload.Channel.ID, payload.Message.Timestamp)
	modalRequest.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			menuList,
		},
	}

	handler.Client.OpenView(payload.TriggerID, modalRequest)
}

// OrderForOther handles when user clicks order for other button
func OrderForOther(handler *Handler, payload *slack.InteractionCallback) {
	message := GetMessageFromTimeStamp(handler.Client, payload.Channel.ID, payload.Message.Timestamp)
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)

	// Menu Select Block
	menuSelectText := slack.NewTextBlockObject("plain_text", "메뉴를 고르라옹", false, false)
	menuSelectElement := slack.NewOptionsSelectBlockElement("static_select", nil, ids.SubmitMenuInput, menuBoard.ToOptionBlockObjects()...)
	menuSelect := slack.NewInputBlock(ids.SubmitMenuInputBlock, menuSelectText, menuSelectElement)

	// User Select Block
	userSelectText := slack.NewTextBlockObject("plain_text", "메뉴를 선택/취소할 사람들도 고르라옹", false, false)
	multiUserSelect := slack.NewOptionsMultiSelectBlockElement("multi_users_select", nil, ids.SubmitMenuPeople)
	userSelect := slack.NewInputBlock(ids.SubmitMenuSelectPeopleBlock, userSelectText, multiUserSelect)

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = slack.NewTextBlockObject("plain_text", "메뉴 선택/취소 대신해주기", false, false)
	modalRequest.Close = slack.NewTextBlockObject("plain_text", "Close", false, false)
	modalRequest.Submit = slack.NewTextBlockObject("plain_text", "Submit", false, false)
	modalRequest.CallbackID = ids.SubmitOrderForOtherCallback
	modalRequest.PrivateMetadata = WriteCallbackMetadata(payload.Channel.ID, payload.Message.Timestamp)
	modalRequest.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			menuSelect, userSelect,
		},
	}

	handler.Client.OpenView(payload.TriggerID, modalRequest)
}

// TerminateMenu handles when user clicks terminate button
func TerminateMenu(handler *Handler, payload *slack.InteractionCallback) {
	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()

	message := GetMessageFromTimeStamp(handler.Client, payload.Channel.ID, payload.Message.Timestamp)
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)

	if message.ParentUserId != payload.User.ID {
		return
	}

	blocks := []slack.Block{}
	for _, curBlock := range message.Blocks.BlockSet {
		switch curBlock.BlockType() {
		case slack.MBTAction:
		case slack.MBTSection:
			curBlock.(*slack.SectionBlock).Accessory = nil
			blocks = append(blocks, curBlock)
		default:
			blocks = append(blocks, curBlock)
		}
	}

	summary := ""
	for _, menu := range menuBoard.Menus {
		choosers := menu.GetChoosers()
		summary += fmt.Sprintf("*%s*\n>", menu.MenuName)
		summary += "`" + strings.Join(choosers, "` `") + "`\n"
	}
	blocks = append(blocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", summary, false, false), nil, nil))

	quote := Quotes[rand.Intn(len(Quotes))]
	blocks = append(blocks, slack.NewDividerBlock(), slack.NewContextBlock(ids.QuoteBlock, slack.NewTextBlockObject("plain_text", quote, false, false)))

	handler.Client.UpdateMessage(payload.Channel.ID, message.Timestamp, slack.MsgOptionBlocks(blocks...))
}

// SelectMenuByUser handles when user select a menu
func SelectMenuByUser(handler *Handler, payload *slack.InteractionCallback, selectedMenuName string) {
	profile, _ := handler.Client.GetUserProfile(&slack.GetUserProfileParameters{UserID: payload.User.ID})

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()

	message := GetMessageFromTimeStamp(handler.Client, payload.Channel.ID, payload.Message.Timestamp)
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)
	menuBoard.ToggleMenuByUser(profile, selectedMenuName)
	handler.Client.UpdateMessage(payload.Channel.ID, message.Timestamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
}

// SubmitMenuAdd handles when user submit menu add view
func SubmitMenuAdd(handler *Handler, payload *slack.InteractionCallback) {
	channel, originalPostTimeStamp := ParseCallbackMetadata(payload.View.PrivateMetadata)

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()

	message := GetMessageFromTimeStamp(handler.Client, channel, originalPostTimeStamp)
	menuName := payload.View.State.Values[ids.SubmitMenuInputBlock][ids.SubmitMenuInput].Value
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)
	emoji, _ := handler.EmojiManager.GetRandomEmoji()
	menuBoard.AddMenu(menuName, emoji)

	// Select default selected users
	selectedUsers := payload.View.State.Values[ids.SubmitMenuSelectPeopleBlock][ids.SubmitMenuPeople].SelectedUsers
	for _, user := range selectedUsers {
		profile, _ := handler.Client.GetUserProfile(&slack.GetUserProfileParameters{UserID: user})
		menuBoard.ToggleMenuByUser(profile, menuName)
	}
	handler.Client.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
}

// SubmitOrderForOther handles when user submit order for other view
func SubmitOrderForOther(handler *Handler, payload *slack.InteractionCallback) {
	channel, originalPostTimeStamp := ParseCallbackMetadata(payload.View.PrivateMetadata)

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()

	message := GetMessageFromTimeStamp(handler.Client, channel, originalPostTimeStamp)
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)
	menuName := payload.View.State.Values[ids.SubmitMenuInputBlock][ids.SubmitMenuInput].SelectedOption.Value
	selectedUsers := payload.View.State.Values[ids.SubmitMenuSelectPeopleBlock][ids.SubmitMenuPeople].SelectedUsers

	// Select selected users
	for _, user := range selectedUsers {
		profile, _ := handler.Client.GetUserProfile(&slack.GetUserProfileParameters{UserID: user})
		menuBoard.ToggleMenuByUser(profile, menuName)
	}

	handler.Client.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
}

// SubmitMenuDelete handles when user submit menu delete view
func SubmitMenuDelete(handler *Handler, payload *slack.InteractionCallback) {
	channel, originalPostTimeStamp := ParseCallbackMetadata(payload.View.PrivateMetadata)

	messageUpdateMutex.Lock()
	defer messageUpdateMutex.Unlock()

	message := GetMessageFromTimeStamp(handler.Client, channel, originalPostTimeStamp)
	menuName := payload.View.State.Values[ids.SubmitMenuDeleteBlock][ids.SubmitMenuInput].SelectedOption.Value
	menuBoard := ParseMenuBlocks(message.Blocks.BlockSet)
	menuBoard.DeleteMenu(menuName)

	handler.Client.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(menuBoard.ToBlocks()...))
}
