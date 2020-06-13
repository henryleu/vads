// +build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"time"

	test "github.com/henryleu/vads/test"
	vad "github.com/henryleu/vads/vad"
	wav "github.com/henryleu/vads/wav"
)

func main() {
	fn := "../data/8ef79f2695c811ea.wav"
	// fn := "../data/16khz-16bits-5.wav"

	r, err := wav.NewReaderFromFile(fn)
	if err != nil {
		log.Fatalf("wav.NewReader() error = %v", err)
	}
	test.InitSpeaker(int(r.FmtChunk.Data.SamplesPerSec), 100)

	c := vad.NewDefaultConfig()
	c.SampleRate = int(r.FmtChunk.Data.SamplesPerSec)
	c.BytesPerSample = int(r.FmtChunk.Data.BitsPerSamples / 8)
	// 设置一下参数效果最佳
	c.SilenceTimeout = 800
	c.SpeechTimeout = 800
	c.NoinputTimeout = 20000
	c.VADLevel = 3

	err = c.Validate()
	if err != nil {
		log.Fatalf("Config.Validate() error = %v", err)
	}
	d := c.NewDetector()
	err = d.Init()
	if err != nil {
		log.Fatalf("Detector.Init() error = %v", err)
	}

	frame := make([]byte, d.BytesPerFrame())
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

events_loop:
	for e := range d.Events {
		switch e.Type {
		// case vad.EventVoiceBegin:
		// 	fmt.Println("voice begin")
		// 	break
		case vad.EventVoiceEnd:
			fmt.Println("voice end")
			f, err := test.NewFile()
			e.Clip.SaveToWriter(f)
			wn := f.Name()
			fmt.Println("name: ", wn)
			rf, err := test.OpenFile(wn)
			if err != nil {
				log.Fatalf("fs.Open() error = %v", err)
			}
			test.PlayWaveFile(rf)
			break events_loop
		case vad.EventNoinput:
			fmt.Println("no input")
			f, err := test.NewFile()
			e.Clip.SaveToWriter(f)
			wn := f.Name()
			fmt.Println("name: ", wn)
			rf, err := test.OpenFile(wn)
			if err != nil {
				log.Fatalf("fs.Open() error = %v", err)
			}
			test.PlayWaveFile(rf)
			break events_loop
		default:
			fmt.Printf("illegal event type %v\n", e.Type)
		}
	}

	time.Sleep(time.Second * 1)
}
