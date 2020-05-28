// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"time"

// 	"github.com/cryptix/wav"
// 	"github.com/faiface/beep"
// 	"github.com/faiface/beep/speaker"
// 	bwav "github.com/faiface/beep/wav"
// 	"github.com/maxhawkins/go-webrtcvad"
// 	"github.com/spf13/afero"
// )

// const kb = 1024
// const mb = 1024 * 1024

// var fs = afero.NewMemMapFs()

// func main2() {
// 	filename := "../data/16khz-16bits-1.wav"

// 	info, err := os.Stat(filename)
// 	checkErr(err)
// 	fmt.Printf("%#v\n", info)

// 	file, err := os.Open(filename)
// 	checkErr(err)
// 	defer file.Close()

// 	wavReader, err := wav.NewReader(file, info.Size())
// 	checkErr(err)

// 	reader, err := wavReader.GetDumbReader()
// 	checkErr(err)

// 	wavInfo := wavReader.GetFile()
// 	fmt.Printf("%#v\n", wavInfo)
// 	rate := int(wavInfo.SampleRate)
// 	if wavInfo.Channels != 1 {
// 		log.Fatal("expected mono file")
// 	}
// 	if rate != 16000 {
// 		log.Fatal("expected 16kHz file")
// 	}

// 	vad, err := webrtcvad.New()
// 	checkErr(err)

// 	err = vad.SetMode(0)
// 	checkErr(err)

// 	frameLength := 160
// 	frameDepth := 2
// 	frame := make([]byte, frameLength*frameDepth)

// 	// if ok := vad.ValidRateAndFrameLength(rate, len(frame)); !ok {
// 	// 	log.Fatal("invalid rate or frame length")
// 	// }

// 	var isActive bool
// 	var offset int

// 	fmt.Println()
// 	// buf := NewBuffer(mb)
// 	// bio := bufio.NewReadWriter(bufio.NewReader(buf), bufio.NewWriter(buf))
// 	// processedFile, err := fs.Create("/tmp/vad.wav")

// 	wf, err := NewWaveFile(fs)
// 	checkErr(err)
// 	meta := wav.File{
// 		Channels:        1,
// 		SampleRate:      wavInfo.SampleRate,
// 		SignificantBits: wavInfo.SignificantBits,
// 	}
// 	w, err := meta.NewWriter(wf)
// 	checkErr(err)

// 	for {
// 		_, err := io.ReadFull(reader, frame)
// 		if err == io.EOF || err == io.ErrUnexpectedEOF {
// 			break
// 		}
// 		checkErr(err)

// 		frameActive, err := vad.Process(rate, frame)
// 		checkErr(err)

// 		if frameActive {
// 			fn := len(frame)
// 			n, err := w.Write(frame)
// 			checkErr(err)
// 			if n != fn {
// 				log.Fatal("fail to write frame to vad file")
// 			}
// 		}
// 		report(frameActive, offset, rate)
// 		isActive = frameActive
// 		// if isActive != frameActive || offset == 0 {
// 		// 	isActive = frameActive
// 		// 	report(isActive, offset, rate)
// 		// }

// 		offset += len(frame)
// 	}

// 	wn := wf.Name()
// 	fmt.Println("name: ", wn)
// 	rf, err := fs.Open(wn)
// 	checkErr(err)
// 	playFile(rf)

// 	report(isActive, offset, rate)
// 	err = w.Close()
// 	checkErr(err)
// }

// // NewBuffer is
// func NewBuffer(cap int) *bytes.Buffer {
// 	buf := make([]byte, 0, cap)
// 	return bytes.NewBuffer(buf)
// }

// // NewWaveWriter is
// func NewWaveWriter(Channels uint16, SampleRate uint32, SignificantBits uint16, fs afero.Fs) (*wav.Writer, error) {
// 	meta := &wav.File{
// 		Channels:        Channels,
// 		SampleRate:      SampleRate,
// 		SignificantBits: SignificantBits,
// 	}

// 	dir := "vad"
// 	pattern := "vad-"
// 	f, err := afero.TempFile(fs, dir, pattern)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return meta.NewWriter(f)
// }

// // NewWaveFile is
// func NewWaveFile(fs afero.Fs) (afero.File, error) {
// 	dir := "vad"
// 	pattern := "vad-"
// 	return afero.TempFile(fs, dir, pattern)
// }

// func checkErr(err error) {
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func report(isActive bool, offset, rate int) {
// 	t := time.Duration(offset) * time.Second / time.Duration(rate) / 2
// 	// fmt.Printf("isActive = %v, t = %v\n", isActive, t)
// 	active := 0
// 	if isActive {
// 		active = 1
// 	} else {
// 		active = 0
// 	}
// 	fmt.Printf("active = %d, offset = %06d, t = %v\n", active, offset, t)
// }

// func playFile(f afero.File) {
// 	streamer, format, err := bwav.Decode(f)
// 	if err != nil {
// 		log.Fatal(fmt.Sprintf("wav.Decode(), error: %v", err))
// 	}
// 	// checkErr(err)
// 	defer streamer.Close()

// 	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
// 	done := make(chan bool)
// 	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
// 		done <- true
// 	})))

// 	<-done
// }
//
