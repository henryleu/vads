package app

import (
	"fmt"
	"io"
	"log"

	ws "github.com/gorilla/websocket"
)

// Conn wraps a net.Conn like Conn
type Conn struct {
	Conn *ws.Conn
	r    io.Reader
}

// NewConn creates Conn between game server and team client.
func NewConn(conn *ws.Conn) *Conn {
	return &Conn{
		Conn: conn,
	}
}

func (c *Conn) ensureReader() error {
	if c.r != nil {
		return nil
	}
	t, r, err := c.Conn.NextReader()
	if t != ws.BinaryMessage {
		log.Printf("conn - ws conn got unwanted reader type: %v\n", t)
	}
	if err != nil {
		return err
	}
	c.r = r
	return nil
}

// Send sends message to the Conn
func (c *Conn) Read(b []byte) (int, error) {
	if err := c.ensureReader(); err != nil {
		return 0, err
	}

	n, err := c.r.Read(b)
	if err == io.EOF || err == ErrUnexpectedEOF

	if c.r == nil {0
		t, r, err := c.Conn.NextReader()
		if t != ws.BinaryMessage {
			log.Printf("ws reader type is %v\n", t)
		}
		if err != nil {
			return 0, err
		}
	}

	wc, err := w.Conn.NextWriter(ws.BinaryMessage)
	if err != nil {
		return fmt.Errorf("Conn error - fail to write message to Conn, error: %v", err)
	}

	if _, err = wc.Write(ConnBytes); err != nil {
		return fmt.Errorf("Conn error - fail to write message to Conn, error: %v", err)
	}
	if debugMessage {
		log.Printf("message sent ->\n%v", string(ConnBytes))
	}
	return wc.Close()
}
