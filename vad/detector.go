package vad

import (
	vad "github.com/maxhawkins/go-webrtcvad"
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

	// NoInputTimeout is no input timeout. By default, 5000 (ms)
	NoInputTimeout int

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

	// FrameTimeBase defines Codec frame time base in msec
	FrameTimeBase int

	// vad is WebRTC VAD processor
	vad *vad.VAD

	sampleCount    int
	vadSampleCount int
	speechDuration int

	debug bool
	work  bool
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
)

// DefaultDetector is
var defaultDetector = Detector{
	State:               StateInactivity,
	SpeechTimeout:       300,
	SilenceTimeout:      300,
	NoInputTimeout:      5000,
	Duration:            0,
	NoInputDuration:     0,
	RecognitionTimeout:  20000,
	RecognitionDuration: 0,
	VADMode:             VADVeryAggressive,
	SampleRate:          SampleRate8,
	BytesPerSample:      BytesPerSample,
	BitsPerSample:       BitsPerSample,
	FrameTimeBase:       FrameTimeBase,
}

// NewDetector creates
func NewDetector() *Detector {
	detector := defaultDetector

	return &detector
}
