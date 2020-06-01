package vad

import (
	"fmt"

	webrtcvad "github.com/maxhawkins/go-webrtcvad"
)

// State is the detector's status during voice activity detecting
type State int

const (
	// StateInactivityTransition means activity detection in-progress
	StateInactivityTransition State = iota

	// StateInactivity means inactivity detected
	StateInactivity

	// StateActivityTransition means inactivity detection is in-progress
	StateActivityTransition

	// StateActivity means activity detected
	StateActivity
)

// Event defines events of activity detector
type Event int

const (
	// EventNone means no event occurred
	EventNone Event = iota

	// EventActivity means voice activity  (transition to activity from inactivity state)
	EventActivity

	// EventInactivity means voice inactivity (transition to activity from inactivity state)
	EventInactivity

	// EventNoInput means no input event occurred
	EventNoInput
)

// Detector detects voice from voice stream based on FSM (finite state machine)
// and VAD library ported from WebRTC
type Detector struct {
	// State is the state of the detector. By default, StateInactivity.
	State State

	// LevelThreshold is the level threshold of voice activity (silence).
	// It should be in [0 .. 255]
	// LevelThreshold int

	// SpeechTimeout is period of activity required to complete transition
	// to active state. By default, 300 (ms)
	SpeechTimeout int

	// SilenceTimeout is period of inactivity required to complete transition
	// to inactive state. By default, 300 (ms)
	SilenceTimeout int

	// Duration is the duration spent in current state. By default, 0
	Duration int

	// NonputTimeout is no input timeout. By default, 5000 (ms)
	NonputTimeout int

	// NoInputDuration is the duration spent during no input state (inactivity state).
	// By default, 0 (ms)
	NoInputDuration int

	// RecognitionTimeout is recognition timeout. By default, 20000 (ms)
	RecognitionTimeout int

	// RecognitionDuration is the duration spent during activity and inactivity transition state.
	// By default, 0 (ms)
	RecognitionDuration int

	// VADMode is the aggressiveness mode for vad. By default, 3 for anti background noise
	VADMode Mode

	// TimersStarted is a flag indicates if timer is on. By default, true
	// TimersStarted bool

	// SampleRate defines the number of samples per second, aka. sample rate.
	// It only supports 8000 and 16000.
	SampleRate int

	// BytesPerSample defines bytes per sample for linear pcm
	BytesPerSample int

	// BitsPerSample defines bits per sample for linear pcm
	BitsPerSample int

	// FrameTime defines Codec frame time in msec. It should be 10ms, 20ms or 30ms. By default, 20 (ms).
	FrameTime int

	// vad is WebRTC VAD processor
	vad *webrtcvad.VAD

	sampleCount    int
	vadSampleCount int
	// speechDuration int

	debug bool
	work  bool
}

// DefaultDetector is
var defaultDetector = Detector{
	State:               StateInactivity,
	SpeechTimeout:       300,
	SilenceTimeout:      300,
	NonputTimeout:       5000,
	Duration:            0,
	NoInputDuration:     0,
	RecognitionTimeout:  20000,
	RecognitionDuration: 0,
	VADMode:             VADVeryAggressive,
	SampleRate:          SampleRate8,
	BytesPerSample:      BytesPerSample,
	BitsPerSample:       BitsPerSample,
	FrameTime:           FrameTimeBase * 2,
}

// NewDetector creates
func NewDetector() *Detector {
	detector := defaultDetector
	return &detector
}

// Init initiates vad and check configuration
func (d *Detector) Init() error {
	vad, err := webrtcvad.New()
	if err != nil {
		// todo logging and wrap error
		return err
	}

	if d.VADMode != VADNormal && d.VADMode != VADLowBitrate && d.VADMode != VADAggressive && d.VADMode != VADVeryAggressive {
		// todo logging and wrap error
		return fmt.Errorf("Detector.VADMode should be 0, 1, 2 or 3, got %d", d.VADMode)
	}

	err = vad.SetMode(int(d.VADMode))
	if err != nil {
		// todo logging and wrap error
		return err
	}
	d.vad = vad

	if d.SampleRate != SampleRate8 && d.SampleRate != SampleRate16 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.SampleRate should be 8000 or 16000, got %d", d.SampleRate)
	}

	if d.FrameTime != FrameTime10 && d.FrameTime != FrameTime20 && d.FrameTime != FrameTime30 {
		// todo logging and wrap error
		return fmt.Errorf("Detector.FrameTime should be 10, 20 or 30, got %d", d.FrameTime)
	}

	return nil
}
