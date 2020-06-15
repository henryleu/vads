package app

import (
	"fmt"
	"io"
	"log"

	ws "github.com/gorilla/websocket"
	wav "github.com/henryleu/vads/wav"
)

// ClientRequest is the handler for nlp+vad
func ClientRequest(url, fn string) {
	c, _, err := ws.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	wire := NewWire(c)
	done := make(chan struct{})

	go wire.ClientReceive()

	go func() {
		defer close(done)
	loop_response:
		for {
			select {
			case msg := <-wire.MsgCh:
				switch msg.Type {
				case ResponseType:
					res, err := msg.Response()
					if err != nil {
						log.Printf("fail to get response msg, error = %v\n", err)
						break loop_response
					}
					log.Printf("succeed to get response: %v\n", res)
					break loop_response
				default:
					log.Printf("client can only receive response msg, but got %v message\n", msg.Type)
					break loop_response
				}
			case err := <-wire.ErrCh:
				log.Printf("fail to receive response msg, error = %v\n", err)
				break loop_response
			}
		}
	}()

	req := &Request{
		CID:  "01010101010",
		Rate: "8000",
		Business: &Business{
			UID:      "1331114444 abcd",
			Province: "beijing",
			Channel:  "03",
		},
	}

	msg := req.Message()
	err = wire.Send(msg)
	if err != nil {
		log.Fatalf("Wire.Send(requestMsg) error = %v", err)
	}

	r, err := wav.NewReaderFromFile(fn)
	if err != nil {
		log.Fatalf("wav.NewReaderFromFile() error = %v", err)
	}

	frame := make([]byte, 1280)
	i := 0

send_chunk:
	for {
		_, err := io.ReadFull(r, frame)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Printf("wav file is EOF, error: %v\n", err)
			break
		}
		if err != nil {
			log.Fatalf("fail to read wav file, error = %v", err)
		}

		i++
		chunk := &Chunk{
			CID:  req.CID,
			NO:   fmt.Sprintf("%d", i),
			Data: frame,
		}
		chunk.EncodeAudio()
		msg := chunk.Message()
		err = wire.Send(msg)
		if err != nil {
			log.Fatalf("Wire.Send(chunkMsg) error = %v", err)
		}
		select {
		case <-done:
			break send_chunk
		default:
		}
		log.Printf("chunk no %v\n", i)
	}

	log.Println("client is done")
	closeConn(c)
}

func closeConn(c *ws.Conn) {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err := c.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, "interrupted"))
	if err != nil {
		log.Println("write close:", err)
		return
	}

}
