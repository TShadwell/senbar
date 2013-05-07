package i3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

var (
	magicStringBytes    = []byte("i3-ipc")
	magicStringBytesLen = len(magicStringBytes)
	messageSize         = binary.Size(ipcMessage{})
)

func (c *Conn) Listen() (err error) {
	defer c.conn.Close()
	reader, pos := byteReader{c.conn}, 0
	var bt byte
	//Read the next magic string.
	for {
		bt, err = reader.ReadByte()

		if err != nil {
			return err
		}

		if bt == magicStringBytes[pos] {
			pos++
		} else {
			pos = 0
		}

		if pos == len(magicStringBytes) {

			msgBuf := make([]byte, messageSize)

			_, err = reader.Read(msgBuf)
			if err != nil {
				return err
			}
			msg := new(ipcMessage)

			//tear off the message information.
			binary.Read(bytes.NewBuffer(msgBuf), binary.LittleEndian, msg)

			var Type interface{}

			if (msgBuf[messageSize-1] >> 7) == byte(1) {
				Type = EventType(msg.Type)
			} else {
				Type = ReplyType(msg.Type)
			}

			var payload json.RawMessage = make([]byte, msg.Length)
			_, err = reader.Read(payload)
			err = c.handlePayload(&payload, Type)
			if err != nil {
				return
			}
			pos = 0
		}

	}
}

func (c *Conn) sendMessage(Type uint32, payload ...byte) (err error) {
	if payload == nil {
		payload = []byte{}
	}
	var end = binary.LittleEndian
	buf := new(bytes.Buffer)
	_, err = buf.Write(magicStringBytes)
	if err != nil {
		return
	}
	err = binary.Write(buf, end, uint32(len(payload)))
	if err != nil {
		return
	}
	err = binary.Write(buf, end, Type)
	if err != nil {
		return
	}
	err = binary.Write(buf, end, payload)
	if err != nil {
		return
	}

	_, err = c.conn.Write(buf.Bytes())
	return
}

func (c *Conn) Command(command string) (success bool, err error) {
	err = c.sendMessage(uint32(Message_Command), []byte(command)...)
	if err != nil {
		return
	}
	var response []successReply
	err = json.Unmarshal(*<-c.command, &response)
	if err != nil {
		return
	}
	return response[0].Success, err
}

func (c *Conn) Workspaces() (ws []Workspace, err error) {
	err = c.sendMessage(uint32(Message_Get_Workspaces))
	if err != nil {
		return
	}
	err = json.Unmarshal(*<-c.get_workspaces, &ws)
	return
}

func (c *Conn) Tree() (root TreeNode, err error) {
	err = c.sendMessage(uint32(Message_Get_tree))
	if err != nil {
		return
	}
	err = json.Unmarshal(*<-c.get_tree, &root)
	return
}

func (c *Conn) Marks() (marks []Mark, err error) {
	err = c.sendMessage(uint32(Message_Get_Marks))
	if err != nil {
		return
	}
	err = json.Unmarshal(*<-c.get_marks, &marks)
	return
}

/*
	Sends a subscribe message to i3. The relevant handlers should be set for this
	Conn, or the events will just be passed-over.

	You can use the more abstracted Subscribe calls instead.
*/
func (c *Conn) Subscribe(events ...Event) (success bool, err error) {
	var byt []byte
	byt, err = json.Marshal(events)
	if err != nil {
		return
	}
	err = c.sendMessage(uint32(Message_Subscribe), byt...)
	var response []successReply
	err = json.Unmarshal(*<-c.subscribe, &response)
	if err != nil {
		return
	}
	return response[0].Success, err
}

func (c *Conn) Outputs() (o []Output, err error) {
	err = c.sendMessage(uint32(Message_Get_Outputs))
	if err != nil {
		return
	}
	err = json.Unmarshal(*<-c.get_outputs, &o)
	return
}

func (c *Conn) handlePayload(j *json.RawMessage, typ interface{}) (err error) {
	var outchan chan *json.RawMessage
	switch v := typ.(type) {
	case ReplyType:
		switch v {
		case Reply_Command:
			outchan = c.command
		case Reply_Workspaces:
			outchan = c.get_workspaces
		case Reply_Subscribe:
			outchan = c.subscribe
		case Reply_Outputs:
			outchan = c.get_outputs
		case Reply_Tree:
			outchan = c.get_tree
		case Reply_Marks:
			outchan = c.get_marks
		case Reply_Bar_Config:
			outchan = c.get_bar_config
		case Reply_Version:
			outchan = c.get_version
		default:
			return errors.New("Invalid ReplyType: " + strconv.Itoa(int(v)) + ".")
		}
	case EventType:
		switch v {
		case Event_Workspace:
			var ev workspaceEvent
			err = json.Unmarshal(*j, &ev)
			if err != nil {
				return
			}
			switch ev.Change {
			case "focus":
				if c.Event.Workspace.Focus != nil {
					c.Event.Workspace.Focus(ev.Current, ev.Old)
				}
			case "init":
				if c.Event.Workspace.Init != nil {
					c.Event.Workspace.Init()
				}
			case "empty":
				if c.Event.Workspace.Empty != nil {
					c.Event.Workspace.Empty()
				}
			case "urgent":
				if c.Event.Workspace.Urgent != nil {
					c.Event.Workspace.Urgent()
				}
			default:
				return errors.New("Invalid WorkspaceEvent: " + ev.Change + ".")
			}
		case Event_Output:
			if c.Event.Output != nil {
				c.Event.Output()
			}
		case Event_Mode:
			var ch changeEvent
			err = json.Unmarshal(*j, &ch)
			if err != nil {
				return
			}
			if c.Event.Mode != nil {
				c.Event.Mode(ch.Change)
			}
		case Event_Window:
			var we windowEvent
			err = json.Unmarshal(*j, &we)
			if err != nil {
				return
			}
			if c.Event.Window != nil {
				c.Event.Window(we.Container)
			}

		default:
			return errors.New("Invalid EventType: " + strconv.Itoa(int(v)) + ".")
		}
		return nil
	}
	if outchan != nil {
		outchan <- j
	}
	return
}

func outputString(b []byte, e error) (string, error) {
	return string(b), e
}

func SocketLocation() (loc string, err error) {
	loc, err = outputString(exec.Command("i3", "--get-socketpath").Output())
	if err != nil {
		return
	}
	if loc == "" {
		err = No_Socket_Loc
	}
	loc = strings.TrimSpace(loc)
	return
}

func Connect(socket string) (c *Conn, e error) {
	var rc io.ReadWriteCloser
	c = new(Conn)
	rc, e = net.Dial("unix", socket)
	if e != nil {
		return
	}
	c = Use(rc)
	return
}

func Attach() (c *Conn, e error) {
	var loc string
	loc, e = SocketLocation()
	if e != nil {
		return
	}
	return Connect(loc)
}

func Use(existingConnection io.ReadWriteCloser) (cn *Conn) {
	cn = &Conn{
		conn: existingConnection,
	}
	//Make channels
	for _, v := range []*chan *json.RawMessage{
		&cn.command,
		&cn.get_workspaces,
		&cn.subscribe,
		&cn.get_outputs,
		&cn.get_tree,
		&cn.get_marks,
		&cn.get_bar_config,
		&cn.get_version,
	} {
		(*v) = make(chan *json.RawMessage, 1)
	}
	return cn
}
