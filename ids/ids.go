package ids

// Action IDs
const (
	SelectMenuByUser = "select_menu_by_user"
	SubmitMenuInput  = "submit_menu_input"
	SubmitMenuPeople = "submit_menu_people"
	AddMenu          = "add_menu"
	DeleteMenu       = "delete_menu"
	TerminateMenu    = "terminate_menu"
)

// Block IDs
const (
	SubmitMenuInputBlock        = "submit_menu_input_block"
	SubmitMenuDeleteBlock       = "submit_menu_delete_block"
	SubmitMenuSelectPeopleBlock = "submit_menu_select_people_block"
	MenuButtonsBlock            = "menu_buttons_block"
	MenuSelectContextBlock      = "menu_select_context_block/"
)

// Callback IDs
const (
	SubmitMenuCallback       = "submit_menu_callback"
	SubmitDeleteMenuCallback = "submit_delete_menu_callback"
)
