package service

import (
	"fmt"
	"slack-waiter-bot/ids"
	"strings"

	"github.com/slack-go/slack"
)

const numHeaderBlocks = 1
const numTailBlocks = 2

// Menu means a menu consist of menu select block and selcted status block
type Menu struct {
	MenuName        string
	MenuSelectBlock *slack.SectionBlock
	StatusBlocks    []*slack.ContextBlock
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
	var menuSelectBlock *slack.SectionBlock
	var statusBlocks []*slack.ContextBlock
	for _, menuBlock := range menuBlocks {
		switch menuBlock := menuBlock.(type) {
		case *slack.SectionBlock:
			if menuSelectBlock != nil {
				menuName := strings.TrimRight(strings.TrimLeft(statusBlocks[0].BlockID, ids.MenuSelectContextBlock), "/0")
				menus = append(menus, Menu{
					MenuName:        menuName,
					MenuSelectBlock: menuSelectBlock,
					StatusBlocks:    statusBlocks,
				})
				menuNameIndexMap[menuName] = len(menuNameIndexMap)
			}
			menuSelectBlock = menuBlock
			statusBlocks = []*slack.ContextBlock{}

		case *slack.ContextBlock:
			statusBlocks = append(statusBlocks, menuBlock)
		}
	}

	if menuSelectBlock != nil {
		menuName := strings.TrimRight(strings.TrimLeft(statusBlocks[0].BlockID, ids.MenuSelectContextBlock), "/0")
		menus = append(menus, Menu{
			MenuName:        menuName,
			MenuSelectBlock: menuSelectBlock,
			StatusBlocks:    statusBlocks,
		})
		menuNameIndexMap[menuName] = len(menuNameIndexMap)
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
	selectText := slack.NewTextBlockObject("plain_text", "👆", false, false)
	menuUserSelectBlock := slack.NewSectionBlock(menuText, nil, slack.NewAccessory(slack.NewButtonBlockElement(ids.SelectMenuByUser, menuName, selectText)))
	menuSelectContextBlock := slack.NewContextBlock(ids.MenuSelectContextBlock+menuName+"/0", slack.NewTextBlockObject("plain_text", "0 Selected", false, false))
	mb.Menus = append(mb.Menus, Menu{
		MenuName:        menuName,
		MenuSelectBlock: menuUserSelectBlock,
		StatusBlocks:    []*slack.ContextBlock{menuSelectContextBlock},
	})
	mb.MenuNameIndexMap[menuName] = len(mb.MenuNameIndexMap)
}

// DeleteMenu deletes the menu
func (mb *MenuBoard) DeleteMenu(menuName string) {
	menuIndex, ok := mb.MenuNameIndexMap[menuName]
	if !ok {
		return
	}

	mb.Menus = append(mb.Menus[:menuIndex], mb.Menus[menuIndex+1:]...)
	delete(mb.MenuNameIndexMap, menuName)
	for name, index := range mb.MenuNameIndexMap {
		if index > menuIndex {
			mb.MenuNameIndexMap[name]--
		}
	}

}

// ToggleMenuByUser select or unselect menu
func (mb *MenuBoard) ToggleMenuByUser(profile *slack.UserProfile, menuName string) {
	menuIndex := mb.MenuNameIndexMap[menuName]
	statusBlocks := mb.Menus[menuIndex].StatusBlocks

	isExist := false
	for i, statusBlock := range statusBlocks {
		elements := statusBlock.ContextElements.Elements
		for j, curElement := range elements {
			curElement, ok := curElement.(*slack.ImageBlockElement)
			if !ok {
				continue
			}
			if curElement.AltText == profile.RealName {
				statusBlock.ContextElements.Elements = append(elements[:j], elements[j+1:]...)
				isExist = true
				break
			}
		}
		if isExist {
			for j := i + 1; j < len(statusBlocks); j++ {
				prevElements := statusBlocks[j-1].ContextElements.Elements
				curElements := statusBlocks[j].ContextElements.Elements
				prevElements[len(prevElements)-1] = curElements[0]
				statusBlocks[j].ContextElements.Elements = curElements[1:]
			}
			break
		}
	}

	lastElements := statusBlocks[len(statusBlocks)-1].ContextElements.Elements
	lastElements = lastElements[:len(lastElements)-1]
	if !isExist {
		lastElements = append(lastElements, slack.NewImageBlockElement(profile.Image32, profile.RealName))
	}

	if len(lastElements) == 10 {
		newStatusBlock := slack.NewContextBlock(fmt.Sprintf("%s/%s/%d", ids.MenuSelectContextBlock, menuName, len(statusBlocks)))
		statusBlocks = append(statusBlocks, newStatusBlock)
		lastElements = newStatusBlock.ContextElements.Elements
	}

	selectedDescription := slack.NewTextBlockObject("plain_text", fmt.Sprintf("%d Selected", 10*(len(statusBlocks)-1)+len(lastElements)), false, false)
	statusBlocks[len(statusBlocks)-1].ContextElements.Elements = append(lastElements, selectedDescription)

	newStatusBlocks := []*slack.ContextBlock{}
	for _, statusBlock := range statusBlocks {
		newStatusBlocks = append(newStatusBlocks, slack.NewContextBlock(statusBlock.BlockID, statusBlock.ContextElements.Elements...))
	}
	mb.Menus[mb.MenuNameIndexMap[menuName]].StatusBlocks = newStatusBlocks
}

// ToBlocks make into blocks
func (mb *MenuBoard) ToBlocks() []slack.Block {
	blocks := []slack.Block{}
	blocks = append(blocks, mb.HeaderBlocks...)

	for _, menu := range mb.Menus {
		blocks = append(blocks, menu.MenuSelectBlock)
		for _, statusBlock := range menu.StatusBlocks {
			blocks = append(blocks, statusBlock)
		}
	}
	blocks = append(blocks, mb.TailBlocks...)
	return blocks
}

// ToOptionBlockObjects make into slack option block object from menu names
func (mb *MenuBoard) ToOptionBlockObjects() []*slack.OptionBlockObject {
	menuOptions := []*slack.OptionBlockObject{}
	for _, menu := range mb.Menus {
		optionBlockText := slack.NewTextBlockObject("plain_text", menu.MenuName, false, false)
		optionBlockObject := slack.NewOptionBlockObject(menu.MenuName, optionBlockText, nil)
		menuOptions = append(menuOptions, optionBlockObject)
	}
	return menuOptions
}
