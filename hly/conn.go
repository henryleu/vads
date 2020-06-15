package hly

import (
	"io"

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
	_, r, err := c.Conn.NextReader()
	// if t != ws.BinaryMessage && t != ws.TextMessage {
	// 	log.Printf("conn - ws conn got unwanted reader type: %v\n", t)
	// }
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
	if err != nil || n == 0 {
		c.r = nil
		return c.Read(b)
	}
	return n, nil
}
