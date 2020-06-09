package test

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/spf13/afero"
)

var fs = afero.NewMemMapFs()

// NewWaveFile creates a temp memory wave file
func NewWaveFile() (afero.File, error) {
	dir := "vad"
	pattern := "vad-"
	return afero.TempFile(fs, dir, pattern)
}

// PlayWaveFile plays from the reader of a wave file
func PlayWaveFile(r io.Reader) {
	streamer, format, err := wav.Decode(r)
	if err != nil {
		log.Fatal(fmt.Sprintf("wav.Decode(), error: %v", err))
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}

func main2() {
	f, err := os.Open("./data/16khz-16bits-1.wav")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
