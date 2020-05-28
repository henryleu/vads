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
	wave "github.com/henryleu/vads/wave"
	"github.com/maxhawkins/go-webrtcvad"
	"github.com/spf13/afero"
)

const kb = 1024
const mb = 1024 * 1024

var fs = afero.NewMemMapFs()

func main() {
	filename := "../data/16khz-16bits-1.wav"

	r, err := wave.NewReader(filename)
	checkErr(err)
	fmt.Printf("NumSamples: %d\n", r.NumSamples)
	fmt.Printf("SampleTime: %d\n", r.SampleTime)
	fmt.Printf("FmtChunk.Data.Channel: %v\n", r.FmtChunk.Data.Channel)
	fmt.Printf("FmtChunk.Data.SamplesPerSec: %v\n", r.FmtChunk.Data.SamplesPerSec)
	fmt.Printf("FmtChunk.Data.BytesPerSec: %v\n", r.FmtChunk.Data.BytesPerSec)
	fmt.Printf("FmtChunk.Data.BlockSize: %v\n", r.FmtChunk.Data.BlockSize)
	fmt.Printf("FmtChunk.Data.BitsPerSamples: %v\n", r.FmtChunk.Data.BitsPerSamples)
	checkErr(err)
	defer r.Close()
	return
	rate := int(r.FmtChunk.Data.Rate)
	if wavInfo.Channels != 1 {
		log.Fatal("expected mono file")
	}
	if rate != 16000 {
		log.Fatal("expected 16kHz file")
	}

	vad, err := webrtcvad.New()
	checkErr(err)

	err = vad.SetMode(0)
	checkErr(err)

	frameLength := 160
	frameDepth := 2
	frame := make([]byte, frameLength*frameDepth)

	// if ok := vad.ValidRateAndFrameLength(rate, len(frame)); !ok {
	// 	log.Fatal("invalid rate or frame length")
	// }

	var isActive bool
	var offset int

	fmt.Println()
	// buf := NewBuffer(mb)
	// bio := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
	// processedFile, err := fs.Create("/tmp/vad.wav")

	wf, err := NewWaveFile(fs)
	checkErr(err)
	meta := wav.File{
		Channels:        1,
		SampleRate:      wavInfo.SampleRate,
		SignificantBits: wavInfo.SignificantBits,
	}
	w, err := meta.NewWriter(wf)
	checkErr(err)

	for {
		_, err := io.ReadFull(reader, frame)
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
		// if isActive != frameActive || offset == 0 {
		// 	isActive = frameActive
		// 	report(isActive, offset, rate)
		// }

		offset += len(frame)
	}

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

// NewWaveWriter is
func NewWaveWriter(Channels uint16, SampleRate uint32, SignificantBits uint16, fs afero.Fs) (*wav.Writer, error) {
	meta := &wav.File{
		Channels:        Channels,
		SampleRate:      SampleRate,
		SignificantBits: SignificantBits,
	}

	dir := "vad"
	pattern := "vad-"
	f, err := afero.TempFile(fs, dir, pattern)
	if err != nil {
		return nil, err
	}

	return meta.NewWriter(f)
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
