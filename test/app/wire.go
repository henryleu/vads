package app

import (
	"bufio"
	"fmt"
	"log"

	ws "github.com/gorilla/websocket"
)

var debugMessage = true

// Wire wraps
type Wire struct {
	Conn  *ws.Conn
	MsgCh chan Message
	ErrCh chan error
}

// NewWire creates wire between game server and team client.
func NewWire(conn *ws.Conn) *Wire {
	return &Wire{
		Conn:  conn,
		MsgCh: make(chan Message, 2),
		ErrCh: make(chan error, 10),
	}
}

// Send sends message to the wire
func (w *Wire) Send(msg *Message) error {
	wireBytes, err := msg.BytesOnWire()
	if err != nil {
		return err
	}

	wc, err := w.Conn.NextWriter(ws.BinaryMessage)
	if err != nil {
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}

	if _, err = wc.Write(wireBytes); err != nil {
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}
	if debugMessage {
		log.Printf("message sent ->\n%v", string(wireBytes))
	}
	return wc.Close()
}

// ClientReceive acts as a client to receive message from server on the wire
func (w *Wire) ClientReceive() {
	scanner := bufio.NewScanner(w.Reader)
	scanner.Split(msgSplit)
	for scanner.Scan() {
		if debugMessage {
			log.Printf("message received ->\n%v", string(scanner.Bytes()))
		}

		msg, err := ParseResponseOnWire(scanner.Bytes())
		if err == nil {
			w.MsgCh <- *msg
			continue
		}

		w.ErrCh <- err
	}
}

// ServerReceive acts as a server to receive message from client on the wire
func (w *Wire) ServerReceive() {
	scanner := bufio.NewScanner(w.Reader)
	scanner.Split(msgSplit)
	for scanner.Scan() {
		if debugMessage {
			log.Printf("message received ->\n%v", string(scanner.Bytes()))
		}
		msg, err := ParseChunkOnWire(scanner.Bytes())
		if err == nil {
			w.MsgCh <- *msg
			continue
		}

		msg, err = ParseRequestOnWire(scanner.Bytes())
		if err == nil {
			w.MsgCh <- *msg
			continue
		}

		w.ErrCh <- err
	}
}

func msgSplit(data []byte, atEOF bool) (int, []byte, error) {
	l := len(data)
	if atEOF && l < MsgHeadLen {
		return 0, nil, fmt.Errorf("message error - message is broken, msg:\n%v", string(data))
	}

	size := ParseMsgLen(data[:MsgHeadLen])
	totalSize := int(size + MsgHeadLen)
	if totalSize > l {
		if atEOF {
			return 0, nil, fmt.Errorf("message error - message is broken, msg:\n%v", string(data))
		}
		return 0, nil, nil
	}

	msgBytes := make([]byte, totalSize, totalSize)
	copy(msgBytes, data[:totalSize])
	return totalSize, msgBytes, nil
}
