package hly

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
)

const debugMessage = true

// Time allowed to write a message to the peer.
const writeWait = 10 * time.Second

// Time to wait before force close on connection.
const closeGracePeriod = 2 * time.Second

// Wire wraps
type Wire struct {
	MsgCh chan Message
	ErrCh chan error
	conn  *ws.Conn
	r     io.Reader
}

// NewWire creates wire between game server and team client.
func NewWire(conn *ws.Conn) *Wire {
	return &Wire{
		MsgCh: make(chan Message, 2),
		ErrCh: make(chan error, 10),
		conn:  conn,
		r:     NewConn(conn),
	}
}

// ClientReceive acts as a client to receive message from server on the wire
func (w *Wire) ClientReceive() {
	scanner := bufio.NewScanner(w.r)
	scanner.Split(msgSplit)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		if debugMessage {
			log.Printf("message received ->\n%v", string(bytes))
		}

		msg, err := ParseResponseOnWire(bytes)
		if err == nil {
			w.MsgCh <- *msg
			continue
		}

		w.ErrCh <- err
	}
}

// ServerReceive acts as a server to receive message from client on the wire
func (w *Wire) ServerReceive() {
	scanner := bufio.NewScanner(w.r)
	scanner.Split(msgSplit)
	first := true
	for scanner.Scan() {
		bytes := scanner.Bytes()
		if debugMessage {
			log.Printf("message received ->\n%v", string(bytes))
		}
		if first {
			first = false
			msg, err := ParseRequestOnWire(bytes)
			log.Printf("ServerReceive first:%s", msg)
			if err == nil {
				w.MsgCh <- *msg
				continue
			}
			w.ErrCh <- err
		} else {
			msg, err := ParseChunkOnWire(bytes)
			log.Printf("ServerReceive second:%s", msg)
			if err == nil {
				w.MsgCh <- *msg
				continue
			}
			w.ErrCh <- err
		}
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

// Send sends message to the wire
func (w *Wire) Send(msg *Message) error {
	wireBytes, err := msg.BytesOnWire()
	if debugMessage {
		log.Printf("message sent ->\n%v", string(wireBytes))
	}
	if err != nil {
		return err
	}

	wc, err := w.conn.NextWriter(ws.BinaryMessage)
	if err != nil {
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}

	if _, err = wc.Write(wireBytes); err != nil {
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}
	return wc.Close()
}

// SendCloseMessage sends close message for closing connection gracefully
func (w *Wire) SendCloseMessage(code int, msg string) {
	w.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
	time.Sleep(closeGracePeriod)
}
