package main

import (
	"fmt"
	"io"
	"log"
	"time"

	vad "github.com/henryleu/vads/vad"
	wav "github.com/henryleu/vads/wav"
)

func main() {
	c := vad.NewDefaultConfig()
	err := c.Validate()
	if err != nil {
		log.Fatalf("Config.Validate() error = %v", err)
	}
	d := c.NewDetector()
	err = d.Init()
	if err != nil {
		log.Fatalf("Detector.Init() error = %v", err)
	}
	fn := "../data/8ef79f2695c811ea.wav"
	r, err := wav.NewReader(fn)
	if err != nil {
		log.Fatalf("wav.NewReader() error = %v", err)
	}

	frame := make([]byte, d.BytesPerFrame())

	go func() {
		for e := range d.Events {
			switch e.Type {
			case vad.EventVoiceBegin:
				fmt.Println("voice begin")
				break
			case vad.EventVoiceEnd:
				fmt.Println("voice end")
				break
			case vad.EventNoinput:
				fmt.Println("no input")
				break
			default:
				fmt.Printf("illegal event type %v\n", e.Type)
			}
		}
	}()

	for {
		_, err := io.ReadFull(r, frame)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Println("file is EOF")
			break
		}
		if err != nil {
			log.Fatalf("io.ReadFull() error = %v", err)
		}
		d.Process(frame)
		if !d.Working() {
			log.Println("detector is stopped")
			break
		}
		if err != nil {
			log.Fatalf("Detector.Process() error = %v", err)
		}
	}

	time.Sleep(time.Second * 5)
}
