package i3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
)

type Error uint8

const (
	No_Socket_Loc Error = iota
)

func (e Error) Error() (o string) {
	switch e {
	case No_Socket_Loc:
		o = "Could not get i3 socket."
	}
	return
}
func (btr byteReader) ReadByte() (b byte, e error) {
	bt := make([]byte, 1)
	_, e = btr.Read(bt)
	if e != nil {
		return 0, e
	}
	return bt[0], e
}

type (
	byteReader struct {
		io.Reader
	}

	Event string

	undefinedType []byte
	Conn          struct {
		Event struct{
			Workspace struct{
				Focus func(current, old *TreeNode)
				Init func()
				Empty func()
				Urgent func()
			}
			Output func()
			Mode func(modeName string)
			Window func(parent TreeNode)
		}
		conn io.ReadWriteCloser
		command,
		get_workspaces,
		subscribe,
		get_outputs,
		get_tree,
		get_marks,
		get_bar_config,
		get_version chan *json.RawMessage
	}
	ipcMessage struct {
		Length uint32
		Type   uint32
	}
	MessageType uint8
	ReplyType   uint8
	EventType   uint8
	borderType  uint8
	layoutType  uint8
	Mark string
	Workspace   struct {
		Number uint
		Name   string
		//Whether the workspace is visible onscreen
		Visible bool
		Urgent  bool
		Rect    Rect
		Output  string
	}
	TreeNode struct {
		ID         int
		Name       string
		Border     borderType
		Layout     layoutType
		Percent    *float32
		Rect       Rect
		WindowRect Rect `json:"window_rect"`
		Geometry   Rect
		Window     int
		Urgent,
		Focused bool
		Nodes []TreeNode
	}
	Output struct {
		Name string
		Active bool
		CurrentWorkspace uint `json:"current_workspace"`
		Rect Rect //#rekt
		Width,
		Height uint
	}
	BarConfig struct {
		/* TODO:This */
	}
	successReply struct {
		Success bool
	}
	Rect struct {
		X,
		Y,
		Width,
		Height uint
	}
	changeEvent struct{
		Change string
	}
	workspaceEvent struct{
		changeEvent
		Old,
		Current *TreeNode
	}
	windowEvent struct{
		Container TreeNode
	}
)

const (
	Border_None borderType = iota
	Border_Normal
	Border_Pixel
)

func (t undefinedType) Type() interface{} {
	btr := []byte{
		t[0],
		t[1],
		t[2],
		t[3],
	}
	var i interface{}
	if t[len(t)-1]>>7 == byte(1) {
		//It is an event.
		i = new(EventType)
	} else {
		i = new(ReplyType)
	}
	binary.Read(
		bytes.NewBuffer(btr),
		binary.LittleEndian,
		&i,
	)
	return reflect.Indirect(reflect.ValueOf(i)).Interface()

}

func (b *borderType) UnmarshalJSON(data []byte) (e error) {
	var t borderType
	switch v := strings.Trim(string(data), "\"'"); v {
	case "normal":
		t = Border_None
	case "none":
		t = Border_None
	case "1pixel", "pixel":
		t = Border_Pixel
	default:
		return errors.New("unrecognised bordertype \"" +
			v +
			"\".")
	}
	(*b) = t
	return
}

const (
	Layout_Split_Horizontal layoutType = iota
	Layout_Split_Vertical
	Layout_Stacked
	Layout_Tabbed
	Layout_Dockarea
	Layout_Output
)

func (l *layoutType) UnmarshalJSON(data []byte) (e error) {
	var t layoutType
	switch v := strings.Trim(string(data), "\"'"); v {
	case "splith":
		t = Layout_Split_Horizontal
	case "splitv":
		t = Layout_Split_Vertical
	case "stacked":
		t = Layout_Stacked
	case "tabbed":
		t = Layout_Tabbed
	case "dockarea":
		t = Layout_Dockarea
	case "output":
		t = Layout_Output
	default:
		return errors.New("unrecognised layoutType \"" +
			v +
			"\".")
	}
	(*l) = t
	return
}

const (
	Message_Command MessageType = iota
	Message_Get_Workspaces
	Message_Subscribe
	Message_Get_Outputs
	Message_Get_tree
	Message_Get_Marks
	Message_Get_Bar_Config
	Message_Get_Version
)

const (
	Reply_Command ReplyType = iota
	Reply_Workspaces
	Reply_Subscribe
	Reply_Outputs
	Reply_Tree
	Reply_Marks
	Reply_Bar_Config
	Reply_Version
)

const (
	Event_Workspace EventType = iota
	Event_Output
	Event_Mode
	Event_Window
)

const (
	EventName_Workspace Event = "workspace"
	EventName_Output    Event = "output"
	EventName_Mode      Event = "mode"
	EventName_Window    Event = "window"
)
