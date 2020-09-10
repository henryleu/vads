package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	wav "github.com/faiface/beep/wav"
	wave "github.com/henryleu/vads/wav"
	"github.com/maxhawkins/go-webrtcvad"
	"github.com/spf13/afero"
)

const kb = 1024
const mb = 1024 * 1024

var fs = afero.NewMemMapFs()

func main() {
	// filename := "../data/16khz-16bits-5.wav"
	filename := "../data/8ef79f2695c811ea.wav"

	r, err := wave.NewReader(filename)
	checkErr(err)
	format := r.FmtChunk.Data
	fmt.Printf("NumSamples: %d\n", r.NumSamples)
	fmt.Printf("SampleTime: %d\n", r.SampleTime)
	fmt.Printf("FmtChunk.Data.Channel: %v\n", format.Channel)
	fmt.Printf("FmtChunk.Data.SamplesPerSec: %v\n", format.SamplesPerSec)
	fmt.Printf("FmtChunk.Data.BytesPerSec: %v\n", format.BytesPerSec)
	fmt.Printf("FmtChunk.Data.BlockSize: %v\n", format.BlockSize)
	fmt.Printf("FmtChunk.Data.BitsPerSamples: %v\n", format.BitsPerSamples)

	if format.Channel != 1 {
		log.Fatal("expected mono file")
	}
	// if format.SamplesPerSec != 16000 {
	// 	log.Fatal("expected 16kHz file")
	// }

	vad, err := webrtcvad.New()
	checkErr(err)

	err = vad.SetMode(2)
	checkErr(err)

	frameLength := 80
	frameDepth := 2
	frame := make([]byte, frameLength*frameDepth)
	var isActive bool = true
	var offset int

	wf, err := NewWaveFile(fs)
	checkErr(err)
	param := wave.WriterParam{
		Out:           wf,
		Channel:       int(format.Channel),
		SampleRate:    int(format.SamplesPerSec),
		BitsPerSample: int(format.BitsPerSamples),
	}
	w, err := wave.NewWriter(param)
	checkErr(err)

	rate := int(format.SamplesPerSec)
	for {
		_, err := io.ReadFull(r, frame)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		checkErr(err)

		frameActive, err := vad.Process(rate, frame)
		checkErr(err)

		if frameActive {
			fn := len(frame)
			n, err := w.Write(frame)
			checkErr(err)
			if n != fn {
				log.Fatal("fail to write frame to vad file")
			}
		}
		report(frameActive, offset, rate)
		isActive = frameActive
		if isActive != frameActive || offset == 0 {
			isActive = frameActive
			// break
		}

		offset += len(frame)
	}
	err = w.Close()
	checkErr(err)

	wn := wf.Name()
	fmt.Println("name: ", wn)
	rf, err := fs.Open(wn)
	checkErr(err)
	playFile(rf)

	report(isActive, offset, rate)
	err = w.Close()
	checkErr(err)
}

// NewBuffer is
func NewBuffer(cap int) *bytes.Buffer {
	buf := make([]byte, 0, cap)
	return bytes.NewBuffer(buf)
}

// NewWaveFile is
func NewWaveFile(fs afero.Fs) (afero.File, error) {
	dir := "vad"
	pattern := "vad-"
	return afero.TempFile(fs, dir, pattern)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func report(isActive bool, offset, rate int) {
	t := time.Duration(offset) * time.Second / time.Duration(rate) / 2
	// fmt.Printf("isActive = %v, t = %v\n", isActive, t)
	active := 0
	if isActive {
		active = 1
	} else {
		active = 0
	}
	fmt.Printf("active = %d, offset = %06d, t = %v\n", active, offset, t)
}

func playFile(f afero.File) {
	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(fmt.Sprintf("wav.Decode(), error: %v", err))
	}
	// checkErr(err)
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
