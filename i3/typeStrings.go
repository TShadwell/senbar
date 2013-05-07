package i3

import (
	"fmt"
)

func (r ReplyType) String() (s string) {
	switch r {
	case Reply_Command:
		s = "Reply_Command"
	case Reply_Workspaces:
		s = "Reply_Workspaces"
	case Reply_Subscribe:
		s = "Reply_Subscribe"
	case Reply_Outputs:
		s = "Reply_Outputs"
	case Reply_Tree:
		s = "Reply_Tree"
	case Reply_Marks:
		s = "Reply_Marks"
	case Reply_Bar_Config:
		s = "Reply_Bar_Config"
	case Reply_Version:
		s = "Reply_Version"
	default:
		s = "Unknown ReplyType #" + fmt.Sprint(uint8(r)) + "."
	}
	return
}

func (c MessageType) String() (s string) {
	switch c {
	case Message_Command:
		s = "Message_Command"
	case Message_Get_Workspaces:
		s = "Message_Get_Workspaces"
	case Message_Subscribe:
		s = "Message_Subscribe"
	case Message_Get_Outputs:
		s = "Message_Get_Outputs"
	case Message_Get_tree:
		s = "Message_Get_tree"
	case Message_Get_Marks:
		s = "Message_Get_Marks"
	case Message_Get_Bar_Config:
		s = "Message_Get_Bar_Config"
	case Message_Get_Version:
		s = "Message_Get_Version"
	default:
		s = "Unknown MessageType #" + fmt.Sprint(uint8(c)) + "."
	}
	return
}

func (e EventType) String() (s string) {
	switch e {
	case Event_Workspace:
		s = "Event_Workspace"
	case Event_Output:
		s = "Event_Output"
	case Event_Mode:
		s = "Event_Mode"
	case Event_Window:
		s = "Event_Window"
	default:
		s = "Unknown EventType #" + fmt.Sprint(uint8(e)) + "."
	}
	return
}
