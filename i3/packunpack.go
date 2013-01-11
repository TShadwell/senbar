package v2

import(
	"json"
)

//i3Message is used to unpack the two int32s from
//the socket.
type i3Message struct {
	PayloadLength uint32
	PayloadType   uint32
}

func packi3Message(payload json.RawMessage, messageType requestType) []byte {
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
