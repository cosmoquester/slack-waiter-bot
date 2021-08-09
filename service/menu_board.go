package service

import (
	"fmt"
	"slack-waiter-bot/ids"

	"github.com/slack-go/slack"
)

const numHeaderBlocks = 1
const numTailBlocks = 2

// Menu means a menu consist of menu select block and selcted status block
type Menu struct {
	MenuName        string
	MenuSelectBlock *slack.SectionBlock
	StatusBlock     *slack.ContextBlock
}

// MenuBoard is menu board blocks containing header, menus, tail
type MenuBoard struct {
	HeaderBlocks     []slack.Block
	TailBlocks       []slack.Block
	Menus            []Menu
	MenuNameIndexMap map[string]int
}

// ParseMenuBlocks parses slack menu board blocks into MenuBoard
func ParseMenuBlocks(blocks []slack.Block) *MenuBoard {
	numBlocks := len(blocks)
	headerBlocks := blocks[:numHeaderBlocks]
	menuBlocks := blocks[numHeaderBlocks : numBlocks-numTailBlocks]
	tailBlocks := blocks[numBlocks-numTailBlocks:]

	menus := []Menu{}
	menuNameIndexMap := map[string]int{}
	for i := 0; i < len(menuBlocks); i += 2 {
		menuSelectBlock := menuBlocks[i].(*slack.SectionBlock)
		statusBlock := menuBlocks[i+1].(*slack.ContextBlock)
		menuName := statusBlock.BlockID[len(ids.MenuSelectContextBlock):]
		menus = append(menus, Menu{
			MenuName:        menuName,
			MenuSelectBlock: menuSelectBlock,
			StatusBlock:     statusBlock})
		menuNameIndexMap[menuName] = i / 2
	}

	return &MenuBoard{
		HeaderBlocks:     headerBlocks,
		TailBlocks:       tailBlocks,
		Menus:            menus,
		MenuNameIndexMap: menuNameIndexMap,
	}
}

// AddMenu adds the menu
func (mb *MenuBoard) AddMenu(menuName string, emoji string) {
	menuText := slack.NewTextBlockObject("plain_text", emoji+menuName, true, false)
	selectText := slack.NewTextBlockObject("plain_text", "Select", false, false)
	menuUserSelectBlock := slack.NewSectionBlock(menuText, nil, slack.NewAccessory(slack.NewButtonBlockElement(ids.SelectMenuByUser, menuName, selectText)))
	menuSelectContextBlock := slack.NewContextBlock(ids.MenuSelectContextBlock+menuName, slack.NewTextBlockObject("plain_text", "0 Selected", false, false))
	mb.Menus = append(mb.Menus, Menu{
		MenuName:        menuName,
		MenuSelectBlock: menuUserSelectBlock,
		StatusBlock:     menuSelectContextBlock,
	})
	mb.MenuNameIndexMap[menuName] = len(mb.MenuNameIndexMap)
}

// ToggleMenuByUser select or unselect menu
func (mb *MenuBoard) ToggleMenuByUser(profile *slack.UserProfile, menuName string) {
	statusBock := mb.Menus[mb.MenuNameIndexMap[menuName]].StatusBlock
	elements := statusBock.ContextElements.Elements
	elements = elements[:len(elements)-1]

	isExist := false
	for i, curElement := range elements {
		if curElement.(*slack.ImageBlockElement).AltText == profile.RealName {
			elements = append(elements[:i], elements[i+1:]...)
			isExist = true
			break
		}
	}
	if !isExist {
		elements = append(elements, slack.NewImageBlockElement(profile.Image32, profile.RealName))
	}

	elements = append(elements, slack.NewTextBlockObject("plain_text", fmt.Sprintf("%d Selected", len(elements)), false, false))
	mb.Menus[mb.MenuNameIndexMap[menuName]].StatusBlock = slack.NewContextBlock(statusBock.BlockID, elements...)
}

// ToBlocks make into blocks
func (mb *MenuBoard) ToBlocks() []slack.Block {
	blocks := []slack.Block{}
	blocks = append(blocks, mb.HeaderBlocks...)

	for _, menu := range mb.Menus {
		blocks = append(blocks, menu.MenuSelectBlock, menu.StatusBlock)
	}
	blocks = append(blocks, mb.TailBlocks...)
	return blocks
}
