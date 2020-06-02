package vad

import (
	"fmt"
)

// Config loads config from json config file.
type Config struct {
	// SpeechTimeout is period of activity required to complete transition
	// to active state. By default, 300 (ms)
	SpeechTimeout int ``

	// SilenceTimeout is period of inactivity required to complete transition
	// to inactive state. By default, 300 (ms)
	SilenceTimeout int

	// NoinputTimeout is no input timeout. By default, 5000 (ms)
	NoinputTimeout int

	// NoinputTimers is a flag indicates if timer is on. By default, true
	NoinputTimers bool

	// RecognitionTimeout is recognition timeout. By default, 20000 (ms)
	RecognitionTimeout int

	// VADLevel is the aggressiveness mode for vad. By default, 3 for anti background noise
	VADLevel VADLevel

	// SampleRate defines the number of samples per second, aka. sample rate.
	// It only supports 8000 and 16000.
	SampleRate int

	// BytesPerSample defines bytes per sample for linear pcm
	BytesPerSample int

	// BitsPerSample defines bits per sample for linear pcm
	BitsPerSample int

	// FrameTime defines Codec frame time in msec. It should be 10ms, 20ms or 30ms. By default, 20 (ms).
	FrameTime int

	// Multiple means if the detector is used to detect multiple speeches.
	// true is for processing a record wave file.
	// false is for processing a incoming voice stream.
	Multiple bool
}

// VADLevel is the aggressiveness level for vad and there are only 4 modes supported.
// 0: vad normal;
// 1: vad low bitrate;
// 2: vad aggressive;
// 3: vad very aggressive;
// By default, 3 is used because it is good at anti background noise.
type VADLevel int

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
	NoinputTimeout:     5000,
	NoinputTimers:      true,
	RecognitionTimeout: 20000,
	VADLevel:           VADVeryAggressive,
	SampleRate:         SampleRate8,
	BytesPerSample:     BytesPerSample,
	BitsPerSample:      BitsPerSample,
	FrameTime:          FrameTime20,
	Multiple:           false,
}

// Validate checks the validity of the config
func (c *Config) Validate() error {
	if c.SpeechTimeout <= 0 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.SpeechTimeout should be greater than 0, got %v", c.SpeechTimeout)
	}

	if c.SilenceTimeout <= 0 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.SilenceTimeout should be greater than 0, got %v", c.SilenceTimeout)
	}

	if c.NoinputTimeout <= 0 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.NoinputTimeout should be greater than 0, got %v", c.NoinputTimeout)
	}

	if c.RecognitionTimeout <= 0 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.RecognitionTimeout should be greater than 0, got %v", c.RecognitionTimeout)
	}

	if c.VADLevel != VADNormal && c.VADLevel != VADLowBitrate && c.VADLevel != VADAggressive && c.VADLevel != VADVeryAggressive {
		// todo logging and wrap error
		return fmt.Errorf("Detector.VADLevel should be 0, 1, 2 or 3, got %v", c.VADLevel)
	}

	if c.SampleRate != SampleRate8 && c.SampleRate != SampleRate16 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.SampleRate should be 8000 or 16000, got %v", c.SampleRate)
	}

	if c.BytesPerSample != BytesPerSample {
		// todo logging and wrap error
		return fmt.Errorf("Detector.BytesPerSample should be 2, got %v", c.BytesPerSample)
	}

	if c.BitsPerSample != BitsPerSample {
		// todo logging and wrap error
		return fmt.Errorf("Detector.BitsPerSample should be 16, got %v", c.BitsPerSample)
	}

	if c.FrameTime != FrameTime10 && c.FrameTime != FrameTime20 && c.FrameTime != FrameTime30 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.FrameTime should be 10, 20 or 30, got %v", c.FrameTime)
	}

	return nil
}

// NewDetector creates a detector with the config populated.
func (c *Config) NewDetector() *Detector {
	return NewDetector(c)
}
