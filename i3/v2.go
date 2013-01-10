package i3

import (
	"strings"
	"error"
	"net"
)

const (
	i3MagicString	= "i3-ipc"
	chunkSize		= 1024
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

type i3Error string

func (i i3Error) Error() string{
	return "Package i3 Error: " + string(i)
}

type chainError error

func (c chainError) Error() string{
	return "\n Recieved error: " + chainError.Error()
}

type unknownResponseChannel interface{}

type req struct{
	request requestType
	responseChannel unknownResponseChannel
}

type Listener struct{
	socket net.Conn
	stack	map[responseType][]req
}

func (l *Listener) pile(rt requestType, channel interface{}){
	l.stack[rt] = append(
		l.stack[rt],
		req{
			rt.getResponse(),
			channel,
		},
	)
}

func Connect(i3SocketLocation string) (*Listener, error)

func Attach() (*Listener, error){
	cmd := exec.Command("i3", "--get-socketpath")
	i3SockLocation, err := cmd.Output()
	if err != nil{
		return nil, err
	}

	return Connect(strings.TrimSpace(string(out)))
}

//Call sends a command to i3 and adds this request to the stack, creating a channel
//that can be used to recieve the message from the Listener.
func (l *Listener) Call(request requestType, payload json.RawMessage) (unknownResponseChannel, error){
	var thisChan unknownResponseChannel
	switch request{
			case REQUEST_COMMAND:
				thisChan = make(chan SubscribeReply)
			case GET_WORKSPACES:
				thisChan = make(chan []Workspace)
			case GET_OUTPUTS:
				thisChan = make(chan []Output)
			case GET_TREE:
				thisChan = make(chan TreeNode)
			case GET_MARKS:
				thisChan = make(chan Marks)
			case GET_BAR_CONFIG:
				thisChan = make(chan BarConfig)
			case GET_VERSION:
				thisChan = make(chan Version)
			case SUBSCRIBE:
				thisChan = make(chan EventResponse)
			default:
				return nil, i3Error("Unknown request type!")

	}
	l.pile(request, thisChan)

	_, err := l.sock.Write(packi3Message(payload, request))
	if err != nil{
		return nil, chainError(err)
	}

	return thisChan
}
