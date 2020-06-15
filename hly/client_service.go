package hly

import (
	"fmt"
	"io"
	"log"

	"github.com/gorilla/websocket"
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
	var errMsg string

	go wire.ClientReceive()

	go func() {
		defer func() {
			close(done)
			log.Println("done channel is closed")
		}()
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
					wire.SendCloseMessage(websocket.CloseNormalClosure, "")
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
		errMsg = fmt.Sprintf("wav.NewReaderFromFile() error = %v\n", err)
		log.Print(errMsg)
		wire.SendCloseMessage(websocket.CloseUnsupportedData, errMsg)
		return
	}

	frame := make([]byte, 1280)
	i := 0

send_chunk:
	for {
		select {
		case <-done:
			break send_chunk
		default:
		}
		_, err := io.ReadFull(r, frame)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Printf("wav file is EOF, error: %v\n", err)
			break send_chunk
		}
		if err != nil {
			errMsg = fmt.Sprintf("fail to read wav file, error = %v", err)
			log.Print(errMsg)
			break send_chunk
		}

		i++
		chunk := &Chunk{
			CID:  req.CID,
			NO:   fmt.Sprintf("%d", i),
			Data: frame,
		}
		chunk.EncodeAudio()
		err = wire.Send(chunk.Message())
		if err != nil {
			errMsg = fmt.Sprintf("Wire.Send(chunkMsg) error = %v", err)
			log.Print(errMsg)
			return
		}
		log.Printf("chunk no %v\n", i)
	}

	log.Println("client is done")
	wire.SendCloseMessage(websocket.CloseUnsupportedData, "")
}
