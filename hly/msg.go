package hly

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// MaxJSONLen is the max length of the JSON bytes of a message
const MaxJSONLen = 60000

// MsgHeadLen is the length of the message header on wire
const MsgHeadLen = 2

// MsgLen defines message length
type MsgLen int

// Bytes converts message length to network order bytes
func (m MsgLen) Bytes() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint16(buf, uint16(m))
	return buf[:2]
}

// ParseMsgLen parses bytes to message length
func ParseMsgLen(buf []byte) MsgLen {
	return MsgLen(binary.BigEndian.Uint16(buf))
}

const (
	// RequestType defines the type of request message
	RequestType = "request"

	// ChunkType defines the type of chunk message
	ChunkType = "chunk"

	// ResponseType defines the type of response message
	ResponseType = "response"
)

// Message is the abstract struct type of Request, Response and Chunk
type Message struct {
	Type    string
	Payload Payload
}

// Payload can be Request, Chunk or Response
type Payload interface{}

// ParseRequestOnWire parsea bytes on wire to request
func ParseRequestOnWire(bytes []byte) (*Message, error) {
	var req Request
	buf := bytes[MsgHeadLen:]
	err := json.Unmarshal(buf, &req)
	if err == nil {
		return &Message{
			Type:    RequestType,
			Payload: &req,
		}, nil
	}

	return nil, fmt.Errorf("message error - fail to unmarshal bytes to request, error: %v\n%v", err, string(buf))
}

// ParseChunkOnWire parsea bytes on wire to request
func ParseChunkOnWire(bytes []byte) (*Message, error) {
	var chk Chunk
	buf := bytes[MsgHeadLen:]
	err := json.Unmarshal(buf, &chk)
	if err == nil {
		return &Message{
			Type:    ChunkType,
			Payload: &chk,
		}, nil
	}
	return nil, fmt.Errorf("message error - fail to unmarshal bytes to chunk, error: %v\n%v", err, string(buf))
}

// ParseResponseOnWire parsea bytes on wire to response
func ParseResponseOnWire(bytes []byte) (*Message, error) {
	var res Response
	buf := bytes[MsgHeadLen:]
	err := json.Unmarshal(buf, &res)
	if err == nil {
		return &Message{
			Type:    ResponseType,
			Payload: &res,
		}, nil
	}
	return nil, fmt.Errorf("message error - fail to unmarshal bytes to response, error: %v\n%v", err, string(buf))
}

// BytesOnWire returns the bytes of the messsage on the wire.
func (m *Message) BytesOnWire() ([]byte, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("message error - fail to marshal message to bytes, error: %v", err)
	}

	jsonLen := len(jsonBytes)
	if jsonLen > MaxJSONLen {
		return nil, fmt.Errorf("message error - message length %v is greather than %v", jsonLen, MaxJSONLen)
	}

	bytes := make([]byte, 0, jsonLen+2)
	lenBytes := MsgLen(jsonLen).Bytes()
	// log.Printf("message length: %v\n", jsonLen)
	// log.Printf("message length bytes: %v\n", lenBytes)
	bytes = append(bytes, lenBytes...)
	bytes = append(bytes, jsonBytes...)
	// log.Println("bytes on wire -> \n" + string(bytes))
	return bytes, nil
}

// String returns the JSON string of the messsage.
func (m *Message) String() string {
	jsonBytes, err := json.MarshalIndent(m, "", "  ")
	// jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("message error - fail to marshal message to bytes, error: %v", err)
	}
	return string(jsonBytes)
}

// MarshalJSON marshals message to json in term of message type
func (m Message) MarshalJSON() ([]byte, error) {
	switch m.Type {
	case RequestType:
		req, ok := m.Payload.(*Request)
		if !ok {
			return nil, fmt.Errorf("message error - payload is not a %v", m.Type)
		}
		return json.Marshal(req)
	case ChunkType:
		chk, ok := m.Payload.(*Chunk)
		if !ok {
			return nil, fmt.Errorf("message error - payload is not a %v", m.Type)
		}
		return json.Marshal(chk)
	case ResponseType:
		res, ok := m.Payload.(*Response)
		if !ok {
			return nil, fmt.Errorf("message error - payload is not a %v", m.Type)
		}
		return json.Marshal(res)
	default:
		return nil, fmt.Errorf("message error - illegal message type %v", m.Type)
	}
}

// Request returns the pointer of the un-marshaled Request obj from payload
func (m *Message) Request() (*Request, error) {
	obj, ok := m.Payload.(*Request)
	if !ok {
		return nil, fmt.Errorf("message error - payload is not a %v but %v", RequestType, m.Type)
	}
	return obj, nil
}

// Response returns the pointer of the un-marshaled Response obj from payload
func (m *Message) Response() (*Response, error) {
	obj, ok := m.Payload.(*Response)
	if !ok {
		return nil, fmt.Errorf("message error - payload is not a %v but %v", ResponseType, m.Type)
	}
	return obj, nil
}

// Chunk returns the pointer of the un-marshaled Chunk obj from payload
func (m *Message) Chunk() (*Chunk, error) {
	obj, ok := m.Payload.(*Chunk)
	if !ok {
		return nil, fmt.Errorf("message error - payload is not a %v but %v", ChunkType, m.Type)
	}
	return obj, nil
}
