package vad

import (
	"fmt"
)

// Config loads config from json cofig file.
type Config struct {
	// SpeechTimeout is period of activity required to complete transition
	// to active state. By default, 300 (ms)
	SpeechTimeout int

	// SilenceTimeout is period of inactivity required to complete transition
	// to inactive state. By default, 300 (ms)
	SilenceTimeout int

	// NonputTimeout is no input timeout. By default, 5000 (ms)
	NonputTimeout int

	// RecognitionTimeout is recognition timeout. By default, 20000 (ms)
	RecognitionTimeout int

	// VADMode is the aggressiveness mode for vad. By default, 3 for anti background noise
	VADMode Mode

	// SampleRate defines the number of samples per second, aka. sample rate.
	// It only supports 8000 and 16000.
	SampleRate int

	// BytesPerSample defines bytes per sample for linear pcm
	BytesPerSample int

	// BitsPerSample defines bits per sample for linear pcm
	BitsPerSample int

	// FrameTime defines Codec frame time in msec. It should be 10ms, 20ms or 30ms. By default, 20 (ms).
	FrameTime int
}

// Mode is the aggressiveness mode for vad and there are only 4 modes supported.
// 0: vad normal;
// 1: vad low bitrate;
// 2: vad aggressive;
// 3: vad very aggressive;
// By default, 3 is used because it is good at anti background noise.
type Mode int

const (

	// VADNormal is normal
	VADNormal = 0

	// VADLowBitrate is low bitrate
	VADLowBitrate = 1

	// VADAggressive is aggressive
	VADAggressive = 2

	// VADVeryAggressive is very aggressive
	VADVeryAggressive = 3
)

const (

	// SampleRate8 is for 8KHZ sample rate
	SampleRate8 = 8000

	// SampleRate16 is for 16KHZ sample rate
	SampleRate16 = 16000

	// BytesPerSample defines bytes per sample for linear pcm
	BytesPerSample = 2

	// BitsPerSample defines bits per sample for linear pcm
	BitsPerSample = 16

	// FrameTimeBase defines Codec frame time base in msec
	FrameTimeBase = 10

	// FrameTime10 is 10ms
	FrameTime10 = FrameTimeBase

	// FrameTime20 is 20ms
	FrameTime20 = FrameTimeBase * 2

	// FrameTime30 is 30ms
	FrameTime30 = FrameTimeBase * 3
)

// DefaultConfig is
var defaultConfig = Config{
	SpeechTimeout:      300,
	SilenceTimeout:     300,
	NonputTimeout:      5000,
	RecognitionTimeout: 20000,
	VADMode:            VADVeryAggressive,
	SampleRate:         SampleRate8,
	BytesPerSample:     BytesPerSample,
	BitsPerSample:      BitsPerSample,
	FrameTime:          FrameTimeBase * 2,
}

// LoadConfig loads config from json file.
func LoadConfig(path string) *Config {
	config := defaultConfig
	return &config
}

func (c *Config) init() error {
	if c.VADMode != VADNormal && c.VADMode != VADLowBitrate && c.VADMode != VADAggressive && c.VADMode != VADVeryAggressive {
		// todo logging and wrap error
		return fmt.Errorf("Detector.VADMode should be 0, 1, 2 or 3, got %c", c.VADMode)
	}

	if c.SampleRate != SampleRate8 && c.SampleRate != SampleRate16 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.SampleRate should be 8000 or 16000, got %c", c.SampleRate)
	}

	if c.FrameTime != FrameTime10 && c.FrameTime != FrameTime20 && c.FrameTime != FrameTime30 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.FrameTime should be 10, 20 or 30, got %c", c.FrameTime)
	}

	return nil
}
