package vad

import (
	"errors"
	"fmt"
	"log"

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

const (
	topSampleRatio = 0.8

	bottomSampleRatio = 0.2

	cacheSecond = 16000 // frame cache for one second

	cacheCap = cacheSecond * 10 // frame cache for 10 seconds

	singleClipCap = 1

	multipleCap = 8
)

// Detector detects voice from voice stream based on FSM (finite state machine)
// and VAD library ported from WebRTC
type Detector struct {
	// Config contains the all the parameters for tuning and controling the detector's behaviors
	*Config

	// Events is a channel for eventing
	Events chan *Event

	// duration is the duration spent in current state. By default, 0
	duration int

	// recognitionDuration is the duration spent during activity and inactivity transition state.
	// By default, 0 (ms)
	recognitionDuration int

	// the starting index (0-based) of current speech in the cache
	speechStart int

	// noinputDuration is the duration spent during no input state (inactivity state).
	// By default, 0 (ms)
	noinputDuration int

	// the starting index (0-based) of current noinput in the cache
	noinputStart int

	// state is the state of the detector. By default, StateInactivity.
	state State

	// vad is WebRTC VAD processor
	vad *webrtcvad.VAD

	sampleCount, vadSampleCount int

	// bytes per millisecond is calculated on sample rate and sample depth (bytes per sample)
	bytesPerMillisecond int

	// bytes per frame is calculated on sample rate, sample depth and frame time
	bytesPerFrame int

	// work indicates if the detector's work is over.
	// true is for working.
	// false is for over.
	work bool

	// frame cache for all incoming samples
	cache []byte

	// all the detected speech clips is here when Detector.Config.Multiple is false.
	clips []*Clip

	// clip is the speech clip or noinput clip when Detector.Config.Multiple is false.
	clip *Clip
}

// DefaultDetector is
var defaultDetector = Detector{
	state:               StateInactivity,
	duration:            0,
	recognitionDuration: 0,
	speechStart:         0,
	noinputDuration:     0,
	noinputStart:        0,
	sampleCount:         0,
	vadSampleCount:      0,
	bytesPerMillisecond: 0,
	bytesPerFrame:       0,
	work:                true,
}

// NewDetector creates
func NewDetector(config *Config) *Detector {
	d := defaultDetector
	d.Config = config
	d.Events = make(chan *Event, 0)
	d.cache = make([]byte, 0, cacheCap)
	if d.Multiple {
		d.clips = make([]*Clip, 0, multipleCap)
	} else {
		d.clips = make([]*Clip, 0, singleClipCap)
	}
	return &d
}

// Init initiates vad and check configuration
func (d *Detector) Init() error {
	vad, err := webrtcvad.New()
	if err != nil {
		// todo logging and wrap error
		return err
	}

	err = vad.SetMode(int(d.VADLevel))
	if err != nil {
		// todo logging and wrap error
		return err
	}
	d.vad = vad

	// calc bytes per unit (millisecond and frame)
	d.bytesPerMillisecond = d.SampleRate * d.BytesPerSample / 1000
	d.bytesPerFrame = d.bytesPerMillisecond * d.FrameDuration

	return nil
}

func (d *Detector) setState(state State) {
	d.state = state
	d.duration = 0
}

func (d *Detector) resetNoinput() {
	d.noinputDuration = 0
	d.noinputStart = 0
}

func (d *Detector) startNoinput() {
	d.noinputDuration = d.duration
	d.noinputStart = len(d.cache) - d.duration*d.bytesPerMillisecond
}

func (d *Detector) endNoinput() {
	d.clip = &Clip{
		SampleRate:     d.SampleRate,
		BytesPerSample: d.BytesPerSample,
		Start:          d.noinputStart / d.bytesPerMillisecond,
		Duration:       d.noinputDuration,
		Data:           d.cache[d.noinputStart:],
	}
	d.resetNoinput()
	d.resetSpeech()
	d.work = false
}

func (d *Detector) resetSpeech() {
	d.recognitionDuration = 0
	d.speechStart = 0
}

func (d *Detector) startSpeech() {
	d.resetNoinput()
	d.recognitionDuration = d.duration
	d.speechStart = len(d.cache) - d.duration*d.bytesPerMillisecond
}

func (d *Detector) endSpeech(transDuration int) {
	l := len(d.cache)
	speechEnd := l - transDuration*d.bytesPerMillisecond
	clip := &Clip{
		SampleRate:     d.SampleRate,
		BytesPerSample: d.BytesPerSample,
		Start:          d.speechStart / d.bytesPerMillisecond,
		Duration:       (l-d.speechStart)/d.bytesPerMillisecond - transDuration,
		Data:           d.cache[d.speechStart:speechEnd],
	}
	d.resetSpeech()
	d.resetNoinput()
	if !d.Multiple {
		d.work = false
		d.clip = clip
	} else {
		d.clips = append(d.clips, clip)
	}
}

func (d *Detector) emitVoiceBegin() {
	if d.Multiple {
		return
	}
	d.Events <- &Event{Type: EventVoiceBegin}
}

func (d *Detector) emitVoiceEnd() {
	if d.Multiple {
		return
	}
	d.Events <- &Event{Type: EventVoiceEnd, Clip: d.clip}
}

func (d *Detector) emitNoinput() {
	d.Events <- &Event{Type: EventNoinput, Clip: d.clip}
}

// Process process the frame of incoming voice samples and generate detection event
func (d *Detector) Process(frame []byte) error {
	// check if the detector is still working
	if !d.Multiple && !d.work {
		log.Println("ignore processing the frame since the detector stopped working")
		return nil
	}

	// calc real times in the frame
	l := len(frame)
	if l%d.bytesPerMillisecond != 0 {
		return fmt.Errorf("frame length is exactly divided with bytes per milliseconds, got %v", l)
	}
	if l%d.bytesPerFrame != 0 {
		return fmt.Errorf("frame length is exactly divided with bytes per frame, got %v", l)
	}

	result, err := d.vad.Process(d.SampleRate, frame)
	if err != nil {
		msg := fmt.Sprintf("Fail to vad process - %v", err)
		log.Println(msg)
		return errors.New(msg)
	}
	frameDuration := l / d.bytesPerMillisecond
	d.cache = append(d.cache, frame...)
	d.duration += frameDuration

	// check recognition timeout
	if d.state == StateActivity || d.state == StateInactivityTransition {
		d.recognitionDuration += frameDuration
		if !d.Multiple && d.RecognitionTimers {
			if d.recognitionDuration >= d.RecognitionTimeout {
				d.endSpeech(0)
				d.emitVoiceEnd()
				return nil
			}
		}
	}

	// check noinput timeout
	if d.state == StateInactivity || d.state == StateActivityTransition {
		d.noinputDuration += frameDuration
		if !d.Multiple && d.NoinputTimers {
			if d.noinputDuration >= d.NoinputTimeout {
				d.endNoinput()
				d.emitNoinput()
				return nil
			}
		}
	}

	switch d.state {
	case StateInactivity:
		if result {
			// start to detect activity
			d.sampleCount = 0
			d.vadSampleCount = 0
			d.setState(StateActivityTransition)
		}
		break
	case StateActivityTransition:
		d.sampleCount++
		if result {
			d.vadSampleCount++
		}
		if result || float32(d.vadSampleCount/d.sampleCount) > topSampleRatio {
			if d.duration >= d.SpeechTimeout {
				// finally detected activity
				d.startSpeech()
				d.setState(StateActivity)
				d.emitVoiceBegin()
			}
		} else {
			// fall back to inactivity
			d.setState(StateInactivity)
		}
		break
	case StateActivity:
		if !result {
			// start to detect inactivity
			d.sampleCount = 0
			d.vadSampleCount = 0
			d.setState(StateInactivityTransition)
		}
		break
	case StateInactivityTransition:
		d.sampleCount++
		if result {
			d.vadSampleCount++
		}
		if result && float32(d.vadSampleCount/d.sampleCount) > bottomSampleRatio {
			// fallback to activity
			d.setState(StateActivity)
		} else {
			if d.duration >= d.SilenceTimeout {
				// detected inactivity
				d.endSpeech(d.duration)
				d.startNoinput()
				d.setState(StateInactivity)
				d.emitVoiceEnd()
			}
		}
		break
	}
	return nil
}
