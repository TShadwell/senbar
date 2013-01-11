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

type unknownResponseChannel interface{}

type i3Error string

var convertReq = map[requestType]responseType{
	REQUEST_COMMAND:RESPONSE_COMMAND,
	GET_WORKSPACES:WORKSPACES,
	SUBSCRIBE:SUBSCRIPTION_RESULT,
	GET_OUTPUTS:OUTPUTS,
	GET_TREE:TREE,
	GET_MARKS:MARKS,
	GET_BAR_CONFIG:BAR_CONFIG,
	GET_VERSION:VERSION,
}

func (rt requestType) response() responseType{
	return convertReq[rt]
}

func (i i3Error) Error() string{
	return "Package i3 Error: " + string(i)
}

type chainError error

func (c chainError) Error() string{
	return "\n Recieved error: " + chainError.Error()
}


type req struct{
	request requestType
	responseChannel unknownResponseChannel
}

type Listener struct{
	socket net.Conn
	stack	map[responseType][]req
}

func (l *Listener) pile(rt requestType, channel interface{}){
	request := req{
		rt.response(),
		channel,
	}
	if l.stack[rt] == nil{
		l.stack[rt] = []req{request}
	}
	l.stack[rt] = append(
		l.stack[rt],
		request,
	)
}

func (l *Listener) pop(rt responseType) (ok bool, request requestType, responseChannel unknownResponseChannel){
	if l.stack[rt] == nil{
		return
	} else {
		if len(l.stack[rt]) > 0{
			ok, request, responseChannel = true, l.stack[rt][0].request, l.stack[rt][0].responseChannel
			l.stack[rt] = l.stack[rt][1:]
			return
		}
	}
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

func (l *Listener)listen(){
	buffer := make(
		[]byte,
		chunkSize,
	)
	messageParts := make([]byte, 0)
	for {
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

			//If the whole payload could not yet be contained in messageParts
			if (magicEnd + 8 + payloadLength) > uint64(len(messageParts)) {
				//Chop the un-needed bit off just in case it saves memory
				messageParts = messageParts[start:]
				//Continue to read until we have all of it.
				break
			}

			//Cast the int to responseType
			payloadType := responseType(uint64(msg.PayloadType))

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
				//Unload the payload into the appropriate channel.
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
					case l.pop(RESPONSE_COMMAND):
						if chResponse_command != nil{
							payloadJSON := getPayloadJSON()
							ReplyObject := payloadJSON.(map[string]interface{})
							chResponse_command <- CommandReply{
								ReplyObject["success"].(bool),
							}

						}
					case WORKSPACES:
						if chWorkspaces != nil{
							op := make([]Workspace, 0)
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
