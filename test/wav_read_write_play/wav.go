package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	test "github.com/henryleu/vads/test"
	wav "github.com/henryleu/vads/wav"
)

func main() {
	// fn := "../data/8ef79f2695c811ea.wav"
	fn := "../data/16khz-16bits-5.wav"
	fi, err := os.Stat(fn)
	log.Printf("the size of %v is %v", fn, fi.Size())

	r, err := wav.NewReaderFromFile(fn)
	if err != nil {
		log.Fatalf("wav.NewReaderFromFile() error = %v", err)
	}
	r.FmtChunk.Debug()
	f, err := test.NewFile()
	wn := f.Name()
	fmt.Println("name: ", wn)
	// rf, err := test.OpenFile(fn)
	// if err != nil {
	// 	log.Fatalf("fs.Open() error = %v", err)
	// }
	param := wav.WriterParam{
		Out:           f,
		Channel:       int(r.FmtChunk.Data.Channel),
		SampleRate:    int(r.FmtChunk.Data.SamplesPerSec),
		BitsPerSample: int(r.FmtChunk.Data.BitsPerSamples),
	}
	param.Debug()
	w, err := wav.NewWriter(param)
	if err != nil {
		log.Printf("Fail to create a new wave clip writer, error: %v", err)
	}

	frame := make([]byte, r.FmtChunk.Data.BlockSize)

	for {
		_, err := io.ReadFull(r, frame)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Println("file is EOF")
			log.Println(frame)
			break
		}
		if err != nil {
			log.Fatalf("io.ReadFull() error = %v", err)
		}

		_, err = w.Write(frame)
		if err != nil {
			log.Printf("Fail to write wave data, error: %v", err)
		}
	}
	w.Close()
	fi, err = test.FileInfo(wn)
	log.Printf("the size of %v is %v", fn, fi.Size())

	rf, err := test.OpenFile(wn)
	if err != nil {
		log.Fatalf("fs.Open() error = %v", err)
	}
	defer rf.Close()
	test.InitSpeaker(int(r.FmtChunk.Data.SamplesPerSec), 100)
	test.PlayWaveFile(rf)

	time.Sleep(time.Second * 1)
}
