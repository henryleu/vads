package vad

import (
	"log"
	"os"

	"github.com/zenwerk/go-wave"
)

// Clip defines voice clip for processing and persisting
type Clip struct {
	// SampleRate defines the number of samples per second, aka. sample rate.
	SampleRate int

	// BytesPerSample defines bytes per sample (sample depth) for linear pcm
	BytesPerSample int

	// Time defines the starting time of the voice clip in the whole voice
	// stream in milliseconds.
	Start int

	// Duration defines the time span of the voice clip in milliseconds.
	Duration int

	// Data is the chunk data of the voice clip as the specific sample rate and depth
	Data []byte
}

// SaveToWave creates a file and write the clip to a wave file.
func (c *Clip) SaveToWave(path string) error {
	f, err := os.Create(path)
	if err != nil {
		log.Printf("Fail to save clip to %v, error: %v", path, err)
		return err
	}
	param := wave.WriterParam{
		Out:           f,
		Channel:       int(1),
		SampleRate:    int(c.SampleRate),
		BitsPerSample: int(c.BytesPerSample * 8),
	}
	w, err := wave.NewWriter(param)
	defer w.Close()
	if err != nil {
		log.Printf("Fail to save clip to %v, error: %v", path, err)
		return err
	}
	return nil
}
