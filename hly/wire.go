package hly

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	ws "github.com/gorilla/websocket"
)

const debugMessage = false

// Time allowed to write a message to the peer.
const writeWait = 10 * time.Second

// Time to wait before force close on connection.
const closeGracePeriod = 2 * time.Second

// Wire wraps
type Wire struct {
	MsgCh  chan Message
	ErrCh  chan error
	conn   *ws.Conn
	closed bool
	mutex  sync.Mutex
}

// NewWire creates wire between game server and team client.
func NewWire(conn *ws.Conn) *Wire {
	return &Wire{
		MsgCh: make(chan Message, 2),
		ErrCh: make(chan error, 10),
		conn:  conn,
	}
}

// ClientReceive acts as a client to receive message from server on the wire
func (w *Wire) ClientReceive() {
	for {
		_, bytes, err := w.conn.ReadMessage()
		if err != nil {
			if debugMessage {
				log.Println("read:", err)
			}
			w.mutex.Lock()
			w.closed = true
			w.mutex.Unlock()
			break
		}
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
	first := true
	for {
		_, bytes, err := w.conn.ReadMessage()
		if err != nil {
			if debugMessage {
				log.Println("read:", err)
			}
			w.mutex.Lock()
			w.closed = true
			w.mutex.Unlock()
			break
		}
		if debugMessage {
			log.Printf("message received ->\n%v", string(bytes))
		}
		if first {
			first = false
			msg, err := ParseRequestOnWire(bytes)
			if err == nil {
				w.MsgCh <- *msg
				continue
			}
			w.ErrCh <- err
		} else {
			msg, err := ParseChunkOnWire(bytes)
			if err == nil {
				w.MsgCh <- *msg
				continue
			}
			w.ErrCh <- err
		}
	}
}

// Send sends message to the wire
func (w *Wire) Send(msg *Message) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.closed {
		return fmt.Errorf("websocket connection is closed")
	}

	wireBytes, err := msg.BytesOnWire()
	if debugMessage {
		log.Printf("message sent ->\n%v", string(wireBytes))
	}
	if err != nil {
		return err
	}

	wc, err := w.conn.NextWriter(ws.BinaryMessage)
	if err != nil {
		w.closed = false
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}

	if _, err = wc.Write(wireBytes); err != nil {
		return fmt.Errorf("wire error - fail to write message to wire, error: %v", err)
	}
	return wc.Close()
}

// SendCloseMessage sends close message for closing connection gracefully
func (w *Wire) SendCloseMessage(code int, msg string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.closed {
		return
	}

	w.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, msg))
	time.Sleep(closeGracePeriod)
}
