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
	SelectMenuByUserID = "select_menu_by_user"
	SubmitMenuID       = "submit_menu"
	AddMenuButtonID    = "add_menu"
	TerminateMenuID    = "terminate_menu"
)

// Block IDs
const (
	SubmitMenuBlockID        = "submit_menu_block"
	MenuButtonsBlockID       = "menu_buttons_block"
	MenuSelectContextBlockID = "menu_select_context_block/"
)

func showAddMenuModal(channelID, originalTs string) slack.ModalViewRequest {
	menuNameText := slack.NewTextBlockObject("plain_text", "MENU", false, false)
	menuNamePlaceholder := slack.NewTextBlockObject("plain_text", "ex) 회전초밥 32pc", false, false)
	menuNameElement := slack.NewPlainTextInputBlockElement(menuNamePlaceholder, SubmitMenuID)
	menuName := slack.NewInputBlock(SubmitMenuBlockID, menuNameText, menuNameElement)

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

	bot, err := api.AuthTest()
	if err != nil {
		_ = fmt.Errorf("INVALID TOKEN ERROR")
		return
	}
	botUserID := bot.UserID

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
				messages, _, _, _ := api.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: ev.Channel, Timestamp: ev.TimeStamp})
				for _, msg := range messages {
					if msg.User == botUserID {
						return
					}
				}

				headerText := slack.NewTextBlockObject("plain_text", "메뉴판", false, false)
				headerBlock := slack.NewHeaderBlock(headerText)

				addMenuBtnTxt := slack.NewTextBlockObject("plain_text", "메뉴 추가", false, false)
				addMenuBtn := slack.NewButtonBlockElement(AddMenuButtonID, AddMenuButtonID, addMenuBtnTxt)
				terminateBtnTxt := slack.NewTextBlockObject("plain_text", "종료", false, false)
				terminateBtn := slack.NewButtonBlockElement(TerminateMenuID, TerminateMenuID, terminateBtnTxt)

				ButtonBlock := slack.NewActionBlock(MenuButtonsBlockID, addMenuBtn, terminateBtn)
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
			for _, blockAction := range payload.ActionCallback.BlockActions {
				switch blockAction.ActionID {
				case AddMenuButtonID:
					api.OpenView(payload.TriggerID, showAddMenuModal(payload.Channel.ID, payload.Message.Timestamp))
				case TerminateMenuID:
					messages, _, _, _ := api.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: payload.Channel.ID, Timestamp: payload.Message.Timestamp})
					for _, msg := range messages {
						if msg.Timestamp != payload.Message.Timestamp {
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
						api.UpdateMessage(payload.Channel.ID, msg.Timestamp, slack.MsgOptionBlocks(blocks...))
					}

				case SelectMenuByUserID:
					profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: payload.User.ID})

					messages, _, _, _ := api.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: payload.Channel.ID, Timestamp: payload.Message.Timestamp})
					for _, msg := range messages {
						if msg.Timestamp != payload.Message.Timestamp {
							continue
						}

						var block *slack.ContextBlock
						var blockID string
						var blockIndex int
						for i, curBlock := range msg.Blocks.BlockSet {
							if curBlock.BlockType() == slack.MBTContext {
								contextBlock := curBlock.(*slack.ContextBlock)
								if contextBlock.BlockID[len(MenuSelectContextBlockID):] == blockAction.Value {
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
						api.UpdateMessage(payload.Channel.ID, msg.Timestamp, slack.MsgOptionBlocks(msg.Blocks.BlockSet...))
					}
				}
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

				menuString := payload.View.State.Values[SubmitMenuBlockID][SubmitMenuID].Value
				menuText := slack.NewTextBlockObject("plain_text", menuString, false, false)
				selectText := slack.NewTextBlockObject("plain_text", "Select", false, false)
				menuUserSelectBlock := slack.NewSectionBlock(menuText, nil, slack.NewAccessory(slack.NewButtonBlockElement(SelectMenuByUserID, menuString, selectText)))
				menuSelectContextBlock := slack.NewContextBlock(MenuSelectContextBlockID+menuString, slack.NewTextBlockObject("plain_text", "0 Selected", false, false))

				// Append New Menu Block
				blocks := make([]slack.Block, len(msg.Blocks.BlockSet)+2)
				copy(blocks, msg.Blocks.BlockSet[:len(msg.Blocks.BlockSet)-2])
				blocks[len(blocks)-4] = menuUserSelectBlock
				blocks[len(blocks)-3] = menuSelectContextBlock
				copy(blocks[len(blocks)-2:], msg.Blocks.BlockSet[len(msg.Blocks.BlockSet)-2:])

				api.UpdateMessage(channel, originalPostTimeStamp, slack.MsgOptionBlocks(blocks...))
			}

		}

	})

	fmt.Println("[INFO] Server listening")
	http.ListenAndServe(":3000", nil)
}
