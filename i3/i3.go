//(c) TNJ Shadwell under the MIT license:
//Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
//to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
//The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. 
//IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

//Package i3 facilitates communication with the i3 window manager by providing channels through which to communicate with it, and some
//helper functions that can be used as examples if lower level interaction is needed.
//
//Recieving events:
//
//	ok := i3.Subscribe(
//		"workspace",
//		"output",
//	)
//
//	if !ok{
//		panic("Unable to subscribe to events!")
//	}
//
//	go (func(){
//		for{
//			<-i3.chWorkspace
//			fmt.Println("Workspace changed.")
//		}
//	})()
//
//	go(func(){
//		for{
//			<-i3.chOutput
//			fmt.Println("Output changed.")
//		}
//	})()
package i3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const (
	i3MagicString = "i3-ipc"
	chunkSize     = 1024
)

type requestType uint8

const (
	REQUEST_COMMAND requestType = iota
	GET_WORKSPACES
	SUBSCRIBE
	GET_OUTPUTS
	GET_TREE
	GET_MARKS
	GET_BAR_CONFIG
	GET_VERSION
)

type responseType uint8

const (
	RESPONSE_COMMAND responseType = iota
	WORKSPACES
	SUBSCRIPTION_RESULT
	OUTPUTS
	TREE
	MARKS
	BAR_CONFIG
	VERSION
)

type eventType uint8

const (
	//Changed desktop
	WORKSPACE eventType = iota
	//Added or removed a display
	OUTPUT
	MODE
)

func (ev eventType) String() string {
	switch ev {
	case WORKSPACE:
		return "workspace"
	case OUTPUT:
		return "output"
	case MODE:
		return "mode"
	}
	panic("Unknown event type '" + strconv.Itoa(int(ev)) + "'.")
}

//responseType returns the corresponding response type for a request type.
func (req requestType) responseType() responseType {
	switch req {
	case REQUEST_COMMAND:
		return RESPONSE_COMMAND
	case GET_WORKSPACES:
		return WORKSPACES
	case SUBSCRIBE:
		return SUBSCRIPTION_RESULT
	case GET_OUTPUTS:
		return OUTPUTS
	case GET_TREE:
		return TREE
	case GET_MARKS:
		return MARKS
	case GET_BAR_CONFIG:
		return BAR_CONFIG
	case GET_VERSION:
		return VERSION
	}
	panic("Unknown request type '" + strconv.Itoa(int(req)) + "' .")
}

//borderType is a simple enum that saves memory.
type borderType uint8

//Border types for the nodes of TREE and other TreeNode based responses.
const (
	//No border
	BORDER_NONE borderType = iota
	//1 pixel border and window title
	BORDER_NORMAL
	//1 pixel border
	BORDER_1PIXEL
)

//layoutType is a simple enum that saves memory.
type layoutType uint8

//Layout types for tree nodes.
const (
	//Horizontal split
	LAYOUT_SPLITH layoutType = iota
	//Vertical split
	LAYOUT_SPLITV
	//Tabbed layout
	LAYOUT_TABBED
	//Node is immune to desktop switches (is a dock).
	LAYOUT_DOCKAREA
	//Is an output (display)
	LAYOUT_OUTPUT
)

func (Resp responseType) String() string {
	switch Resp {
	case RESPONSE_COMMAND:
		return "RESPONSE_COMMAND"
	case WORKSPACES:
		return "WORKSPACES"
	case SUBSCRIPTION_RESULT:
		return "SUBSCRIPTION_RESULT"
	case OUTPUTS:
		return "OUTPUTS"
	case TREE:
		return "TREE"
	case MARKS:
		return "MARKS"
	case BAR_CONFIG:
		return "BAR_CONFIG"
	case VERSION:
		return "VERSION"
	}
	panic("Response type '" + strconv.Itoa(int(Resp)) + "' is invalid.")
}
func (Resp requestType) String() string {
	switch Resp {
	case REQUEST_COMMAND:
		return "command"
	case GET_WORKSPACES:
		return "get_workspaces"
	case SUBSCRIBE:
		return "subscribe"
	case GET_OUTPUTS:
		return "get_outputs"
	case GET_TREE:
		return "get_tree"
	case GET_MARKS:
		return "get_marks"
	case GET_BAR_CONFIG:
		return "get_bar_config"
	case GET_VERSION:
		return "get_version"
	}
	panic("Response type '" + strconv.Itoa(int(Resp)) + "' is invalid.")
}
func (Bor borderType) String() string {
	switch Bor {
	case BORDER_NONE:
		return "none"
	case BORDER_NORMAL:
		return "normal"
	case BORDER_1PIXEL:
		return "1pixel"
	}
	panic("Border type '" + strconv.Itoa(int(Bor)) + "' is invalid.")
}
func (Lay layoutType) String() string {
	switch Lay {
	case LAYOUT_SPLITH:
		return "splith"
	case LAYOUT_SPLITV:
		return "splitv"
	case LAYOUT_TABBED:
		return "tabbed"
	case LAYOUT_DOCKAREA:
		return "dockarea"
	case LAYOUT_OUTPUT:
		return "output"
	}
	panic("Layout type '" + strconv.Itoa(int(Lay)) + "' is invalid.")
}

type CommandReply struct {
	Success bool
}
type SubscribeReply bool
type Marks []string

//BUG(tnjs): barconfig not yet implimented.
type BarConfig struct {
	/*Not yet implemented!*/
}
type Version struct {
	Major,
	Minor,
	Patch uint8
	HumanReadable string "human_readable"
}
type Rectangle struct {
	Height,
	Width,
	X,
	Y uint32
}

//Workspace represents the attributes of one desktop, or workspace in i3.
type Workspace struct {
	Focused,
	Urgent,
	Visible bool
	Name,
	Output string
	//Num is an undocumented feature in i3, and appears to be some sort of \d+
	//regex on desktop names.
	Num  uint
	Rect Rectangle
}

type Workspaces []Workspace

var listening bool

//Output represents the attributes of one output, or display screen in i3.
type Output struct {
	Name string
	Active,
	Primary bool
	CurrentWorkspace *string
	Rect             Rectangle
}

type Outputs []Output

//TreeNode is the self-similar form in which the window tree is provided.
type TreeNode struct {
	Id      uint32
	Name    string
	Border  borderType
	Layout  layoutType
	Percent *float32
	//Absolute relative to top left of desktop
	Rect Rectangle
	//Relative to top left of container
	WindowRect Rectangle "window_rect"
	Urgent,
	Focused bool
	Nodes []TreeNode
}

//EventResponse is for event types such as 'focus'.
type EventResponse string

//i3Message is used to unpack the two int32s from
//the socket.
type i3Message struct {
	PayloadLength uint32
	PayloadType   uint32
}

//Command response channels.
var (
	chResponse_command    chan CommandReply
	chWorkspaces          chan Workspaces
	chSubscription_result chan SubscribeReply
	chOutputs             chan Outputs
	chTree                chan TreeNode
	chMarks               chan Marks
	chBar_config          chan BarConfig
	chVersion             chan Version
)




chResponse_command    chan CommandReply
chWorkspaces          chan Workspaces
chSubscription_result chan SubscribeReply
chOutputs             chan Outputs
chTree                chan TreeNode
chMarks               chan Marks
chBar_config          chan BarConfig
chVersion             chan Version

//Event channels.
var (
	chWorkspace chan EventResponse
	chOutput    chan EventResponse
	chMode      chan EventResponse
)

var chErrors chan error

/*
Prevents the i3 library from panicing on irrecoverable errors, returning a channel
though which the errors are sent instead.

If an error happens once, it may happen again, locking this channel!
*/
func GetErrors() *chan error{
	if chErrors == nil{
		chErrors = make(chan error)
	}
	return &chErrors
}

var i3MagicStringBytes = []byte(i3MagicString)
var i3MagicStringLength = len(i3MagicStringBytes)
var i3SocketConn net.Conn

func makeBorder(borderType string) borderType {
	switch borderType {
	case "none":
		return BORDER_NONE
	case "normal":
		return BORDER_NORMAL
	case "pixel":
		fallthrough
	case "1pixel":
		return BORDER_1PIXEL
	}
	panic("Border type '" + borderType + "' is invalid.")
}
func stripQuotes(inp string) string {
	return strings.Trim(inp, "\"'")
}
func (b *borderType) UnmarshalJSON(x []byte) error {
	*b = makeBorder(stripQuotes(string(x)))
	return nil
}
func (l *layoutType) UnmarshalJSON(x []byte) error {
	*l = makeLayout(stripQuotes(string(x)))
	return nil
}
func makeLayout(layoutType string) layoutType {
	switch layoutType {
	case "splith":
		return LAYOUT_SPLITH
	case "splitv":
		return LAYOUT_SPLITV
	case "tabbed":
		return LAYOUT_TABBED
	case "dockarea":
		return LAYOUT_DOCKAREA
	case "output":
		return LAYOUT_OUTPUT
	}
	panic("Layout type '" + layoutType + "' is invalid.")
}
func packi3Message(payload string, messageType requestType) []byte {
	/*
		By example:

		b'i3-ipc\t\x00\x00\x00\x01\x00\x00\x00COOLBEANS'
		b'i3-ipc\x08\x00\x00\x00\x01\x00\x00\x00COOLBENS'
		b'i3-ipc - magic string
		\x07\x00\x00\x00 -payload length
		\x01\x00\x00\x00 - payload type
		COOLBNS' - payload
	*/
	nPayload := []byte(payload)
	payloadLen := len(nPayload)
	magic := []byte(i3MagicString)
	magicLen := len(magic)
	//Allocate mem for the slice
	// magic length + 4 for payload length
	// + 4 for type
	// + payload length
	msg := make(
		[]byte,
		magicLen+8+payloadLen)
	//copy the magic in
	rolling := 0
	copy(msg, magic)
	rolling += magicLen
	binary.PutUvarint(
		msg[rolling:],
		//PutUvarint only accepts uint64
		//converted from uint32 to prevent silent overflows
		uint64(uint32(len(payload))))
	//The next write will overwrite the extra bytes
	rolling += 4
	binary.PutUvarint(
		msg[rolling:],
		//PutUvarint only accepts uint64
		//converted from uint32 to prevent silent overflows
		uint64(messageType))
	rolling += 4

	copy(
		msg[rolling:],
		nPayload)
	return msg
}

type i3err uint8
const (
	READ_ERROR i3err = iota
	UNPACK_ERROR
	JSON_PROCESS_ERROR
	UNKNOWN_RESPONSE_TYPE
)

var errMap = map[i3err]string{
	READ_ERROR:"Error reading from socket!",
	UNPACK_ERROR:"Error unpacking message!",
	JSON_PROCESS_ERROR:"Error reading JSON!",
	UNKNOWN_RESPONSE_TYPE:"Unknown response type!",
}

type i3ErrEffect uint8
const (
	//Causes this package to stop listening to i3
	STOP i3ErrEffect = iota

	//Causes the package to abandon the current read.
	ABANDON

	//Causes the package to continue and send nil (usually)
	//down channel
	CONTINUE

)


var ErrorEffect = map[i3err]i3ErrEffect{
	READ_ERROR:STOP,
	UNPACK_ERROR:ABANDON,
	//Sends nil down channels
	JSON_PROCESS_ERROR:CONTINUE,
	UNKNOWN_RESPONSE_TYPE:ABANDON,
}

//Returns the corresponding effect of an error
func (i i3err) Effect() i3ErrEffect{
	return ErrorEffect[i]
}

func (i i3err) String()string{
	return errMap[i]
}

func (i i3err) Error() string{
	return i.String()
}

func IsListening() bool{
	return listening
}

var stopListening bool

//If this package is listening, it stops listening,
//else it returns false.
func StopListening()bool{
	if listening{
		stopListening = true
		return true
	}
	return false
}

//If this package is not listening it starts listening.
//else it returns false.
func StartListening() bool{
	if !listening{
		listen()
		return true
	}
	return false
}

func listen() {
	listening = true
	defer func(){
		listening = false
	}()
	buffer := make(
		[]byte,
		chunkSize,
	)
	messageParts := make([]byte, 0)
	for {
		if stopListening{
			stopListening = false
			runtime.Goexit()
		}
		n, err := i3SocketConn.Read(buffer)
		if err != nil {
			if chErrors == nil{
				panic(READ_ERROR)
			} else {
				chErrors <- READ_ERROR
				break
			}
		}
		messageParts = append(messageParts, buffer[:n]...)
		var start int
		//Look for the start of the message, and discard everything before it.
		//If we have not yet recieved the magic string, read until we do.
		//If we don't have the full message head, read until we do also.
		for start,
			//Set variables for the start of the magic string and the end of the magic string
			magicEnd := bytes.Index(messageParts, i3MagicStringBytes),
			uint64(start+i3MagicStringLength);

		//As long as the magic string is present, and the length of the 
		//byte slice is such that it could contain all neccesary information to
		//Process the string
		start != -1 && !((magicEnd + 8) > uint64(len(messageParts)-1));

		//On each loop, reset the start and end position of the magic string
		//to their actual position in the byte slice
		start, magicEnd = bytes.Index(messageParts, i3MagicStringBytes),
			uint64(start+i3MagicStringLength) {
			//Get the payload length and type

			var msg i3Message

			buf := bytes.NewBuffer(messageParts[magicEnd : magicEnd+8])
			errbin := binary.Read(buf, binary.LittleEndian, &msg)
			if errbin != nil {
				if chErrors == nil{
					panic(UNPACK_ERROR)
				} else {
					chErrors <- UNPACK_ERROR
					continue
				}
			}
			payloadLength := uint64(msg.PayloadLength)
			payloadTypeInt := uint64(msg.PayloadType)

			//If the whole payload could not yet be contained in messageParts
			if (magicEnd + 8 + payloadLength) > uint64(len(messageParts)) {
				//Chop the un-needed bit off just in case it saves memory
				messageParts = messageParts[start:]
				//Continue to read until we have all of it.
				break
			}

			//Unload the payload into the appropriate channel.
			payloadType := responseType(payloadTypeInt)
			jsonString := messageParts[magicEnd+8 : magicEnd+8+payloadLength]
			getPayloadJSON := func() interface{} {
				var out interface{}
				if json.Unmarshal(jsonString, &out) != nil {
					if chErrors == nil{
						panic(JSON_PROCESS_ERROR)
					} else {
						chErrors <- JSON_PROCESS_ERROR
					}
				}
				return out
			}
			//Spec says that the highest value bit is set to one if it is an event.
			if messageParts[magicEnd+7]>>7 == byte(1) {
				eventType := eventType(payloadType)
				payloadJSON := getPayloadJSON()
				eventString := EventResponse(payloadJSON.(map[string]interface{})["change"].(string))
				switch eventType {
				case WORKSPACE:
					//A horrible hack to prevent some horrible race conditions
					//These need to be dealt with in time
					if eventString != "init" && eventString != "empty" && chWorkspace != nil {
						chWorkspace <- eventString
					}
				case OUTPUT:
					if chOutput != nil{
						chOutput <- eventString
					}
				case MODE:
					if chMode != nil{
						chMode <- eventString
					}
				default:
					panic("Unknown event type '" + strconv.Itoa(int(eventType)) + "'.")
				}
			} else {
				switch payloadType {
				case RESPONSE_COMMAND:
					if chResponse_command != nil{
						payloadJSON := getPayloadJSON()
						ReplyObject := payloadJSON.(map[string]interface{})
						chResponse_command <- CommandReply{
							ReplyObject["success"].(bool),
						}

					}
				case WORKSPACES:
					if chWorkspaces != nil{
						op := make(Workspaces, 0)
						json.Unmarshal(jsonString, &op)
						chWorkspaces <- op
					}
				case SUBSCRIPTION_RESULT:
					if chSubscription_result != nil{
						payloadJSON := getPayloadJSON()
						ReplyObject := payloadJSON.(map[string]interface{})
						chSubscription_result <- SubscribeReply(ReplyObject["success"].(bool))
					}
				case OUTPUTS:
					if chOutputs != nil{
						cOutputs := make(Outputs, 0)
						json.Unmarshal(jsonString, &cOutputs)
						chOutputs <- cOutputs
					}
				case TREE:
					if chTree != nil{
						var root TreeNode
						json.Unmarshal(jsonString, &root)
						chTree <- root
					}
				case MARKS:
				case BAR_CONFIG:
				case VERSION:
					if chVersion != nil{
						var version Version
						json.Unmarshal(jsonString, &version)
						chVersion <- version
					}
				default:
					if chErrors != nil{
						chErrors <- UNKNOWN_RESPONSE_TYPE
					} else {
						panic("Unknown response type!")
					}
				}
			}
			//Remove the channeled information
			messageParts = messageParts[magicEnd+8+payloadLength:]
		}
	}
}

//Function Nag calls i3-nagbar with the specified arguments.
func Nag(label string, labelAction ...[2]string) error {
	var arg []string
	if len(labelAction) > 0 {
		arg = []string{
			"-m",
			"'" + label + "'",
			"-b",
			"'" + labelAction[0][0] + "'",
			"'" + labelAction[0][1] + "'",
		}
	} else {
		arg = []string{
			"-m",
			"'" + label + "'",
		}
	}
	return exec.Command("i3-nagbar", arg...).Run()
}

func shell(fun, arg string) (string, error) {
	cmd := exec.Command(fun, arg)
	out, err := cmd.Output()
	return string(out), err
}

//Send sends a message to i3 with given payload and requestType.
func Send(payload string, msgType requestType) {
	_, err := i3SocketConn.Write(packi3Message(payload, msgType))
	if err != nil {
		panic("Error writing to i3 socket!")
	}
}

//GetOutputs sends the GET_OUTPUTS signal, waits for reply
func GetOutputs() Outputs {
	Send("", GET_OUTPUTS)
	return <-chOutputs
}

//GetActive outputs sends the GET_OUTPUTS signal, filters out outputs not being used.
func GetActiveOutputs() Outputs {
	outputs := GetOutputs()
	fOutputs := make(Outputs, 0)
	for _, output := range outputs {
		if output.Active {
			fOutputs = append(fOutputs, output)
		}
	}
	return fOutputs
}

//GetWorkspaces returns an array of workspaces (desktops).
func GetWorkspaces() Workspaces {
	Send("", GET_WORKSPACES)

	return <-chWorkspaces
}

//GetTree returns a tree of windows.
func GetTree() TreeNode {
	Send("", GET_TREE)

	return <-chTree
}

//Subscribe -  to a list of i3 events, returns success as bool.
func Subscribe(events ...string) bool {
	val, err := json.Marshal(events)
	if err != nil {
		panic("Marshalling error!")
	}
	Send(
		string(val),
		SUBSCRIBE)
	return bool(<-chSubscription_result)
}

//WorkspacesPerDisplay sorts workspaces by display; useful for status bars.
func WorkspacesPerDisplay() map[string]Workspaces {
	Send("", GET_WORKSPACES)
	workspaces := <-chWorkspaces
	cWorkspaces := make(map[string]Workspaces)

	for _, workspace := range workspaces {
		concernedOutput, ok := cWorkspaces[workspace.Output]
		if ok == false {
			cWorkspaces[workspace.Output] = Workspaces{
				workspace}
		} else {
			cWorkspaces[workspace.Output] = append(concernedOutput, workspace)
		}
	}
	return cWorkspaces
}

//Fail calls Nag(s) and panic(s).
func Fail(s string) {
	Nag(s)
	panic(s)
}

func init() {
	i3SockLoc, err := shell("i3", "--get-socketpath")
	i3SockLoc = strings.TrimSpace(i3SockLoc)
	if err != nil {
		Fail("Unable to get socketpath!")
	}
	conn, erro := net.Dial("unix", i3SockLoc)
	if erro != nil {
		Fail("Unable to connect to i3 socket!")
	}
	i3SocketConn = conn

	go listen()
}
