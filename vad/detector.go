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

const (
	topVolumeValve = 0.8

	bottomVolumeValve = 0.2

	cacheSecond = 16000 // frame cache for one second

	cacheCap = cacheSecond * 10 // frame cache for 10 seconds
)

// Detector detects voice from voice stream based on FSM (finite state machine)
// and VAD library ported from WebRTC
type Detector struct {
	// Config contains the all the parameters for tuning and controling the detector's behaviors
	*Config

	// State is the state of the detector. By default, StateInactivity.
	State State

	// Duration is the duration spent in current state. By default, 0
	Duration int

	// NoinputDuration is the duration spent during no input state (inactivity state).
	// By default, 0 (ms)
	NoinputDuration int

	// RecognitionDuration is the duration spent during activity and inactivity transition state.
	// By default, 0 (ms)
	RecognitionDuration int

	// vad is WebRTC VAD processor
	vad *webrtcvad.VAD

	sampleCount, vadSampleCount int

	speechStart, speechEnd int

	// bytes per millisecond based sample rate and sample depth (bytes per sample)
	bytesPerMillisecond int

	// work indicates if the detector's work is over.
	// true is for working.
	// false is for over.
	work bool

	// frame cache for all incoming samples
	cache []byte
}

// DefaultDetector is
var defaultDetector = Detector{
	State:               StateInactivity,
	Duration:            0,
	NoinputDuration:     0,
	RecognitionDuration: 0,
	sampleCount:         0,
	vadSampleCount:      0,
	speechStart:         0,
	speechEnd:           0,
	work:                true,
}

// NewDetector creates
func NewDetector(config *Config) *Detector {
	d := defaultDetector
	d.Config = config
	d.cache = make([]byte, 0, cacheCap)
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

	d.bytesPerMillisecond = d.SampleRate * d.BitsPerSample / 1000

	// todo init other resources
	return nil
}

func (d *Detector) setState(state State) {
	d.State = state
	d.Duration = 0
}

func (d *Detector) cacheFrame(frame []byte) {
	d.cache = append(d.cache, frame...)
}

// Process process the frame of incoming voice samples and generate detection event
func (d *Detector) Process(frame []byte) error {
	// check if the detector is still working
	if d.Multiple && !d.work {
		return nil
	}

	// todo validate frame

	// calc real times in the frame
	frameTime := len(frame) / d.bytesPerMillisecond

	d.cacheFrame(frame)

	result, err := d.vad.Process(d.SampleRate, frame)
	if err != nil {
		msg := fmt.Sprintf("Fail to vad process - %v", err)
		log.Println(msg)
		return errors.New(msg)
	}

	// todo result and volume level checking

	// check recognition timeout
	if d.State == StateActivity || d.State == StateInactivityTransition {
		d.RecognitionDuration += frameTime
		if d.RecognitionDuration >= d.RecognitionTimeout {
			d.setState(StateInactivity)
			d.RecognitionDuration = 0
			// todo emit event(inactivity)
			return nil
		}
	}

	switch d.State {
	case StateInactivity:
		if result {
			// start to detect activity
			d.sampleCount = 0
			d.vadSampleCount = 0
			d.setState(StateActivityTransition)
		} else {
			if d.NoinputTimers {
				d.Duration += frameTime
				d.NoinputDuration += frameTime
				if d.NoinputDuration >= d.NoinputTimeout {
					// detected noinput
					// todo emit event(noinput)
					d.NoinputDuration = 0
				}
			}
		}
		break
	case StateActivityTransition:
		d.sampleCount++
		if result {
			d.vadSampleCount++
		}
		if result || float32(d.vadSampleCount/d.sampleCount) > topVolumeValve {
			d.Duration += frameTime
			if d.Duration >= d.SpeechTimeout {
				// finally detected activity
				// d.speechDuration = d.Duration
				// todo record speech
				d.setState(StateActivity)
				// todo emit event(activity)
			}
		} else {
			// fallback to inactivity
			d.NoinputDuration += frameTime
			d.setState(StateInactivity)
		}
		break
	case StateActivity:
		if result {
			d.Duration += frameTime
		} else {
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
		if result && float32(d.vadSampleCount/d.sampleCount) > bottomVolumeValve {
			// fallback to activity
			d.setState(StateActivity)
		} else {
			d.Duration += frameTime
			if d.Duration >= d.SilenceTimeout {
				// detected inactivity
				if !d.Multiple {
					d.work = false
				}
				d.setState(StateInactivity)
				// todo emit event(inactivity)
			}
		}

		break
	}
	return nil
	/*

		  mpf_detector_event_e det_event = MPF_DETECTOR_EVENT_NONE;
		  if (detector->work == FALSE) {
		    return det_event;
		  }

		  apt_bool_t result = FALSE;
		  int time_base = CODEC_FRAME_TIME_BASE;
		  apr_size_t level = mpf_activity_detector_level_calculate(frame);
		  int vad = WebRtcVad_Process(detector->vad_inst,
		                              detector->voice_rate,
		                              frame->codec_frame.buffer,
		                              frame->codec_frame.size / 2,
		                              1);

		  if (level >= detector->level_threshold && vad == 1) {
		    result = TRUE;
		  }

		  if (detector->state == DETECTOR_STATE_ACTIVITY || detector->state == DETECTOR_STATE_INACTIVITY_TRANSITION) {
		    detector->recognition_duration += time_base;
		    if (detector->recognition_duration >= detector->recognition_timeout) {
		      mpf_activity_detector_state_change(detector, DETECTOR_STATE_INACTIVITY);
		      detector->recognition_duration = 0;
		      return MPF_DETECTOR_EVENT_INACTIVITY;
		    }
		  }

			if (detector->debug == TRUE) {
		    printf("voice info: level=%5d, google=%d, mrcp=%d, state=%d, duration=%dms\n",
		           (int) level,
		           vad,
		           result,
		           detector->state,
		           (int) detector->recognition_duration);
		  }
	*/

	/*
	  if (detector->state == DETECTOR_STATE_INACTIVITY) {
	    if (result == TRUE) {
	      // start to detect activity
	      detector->sample_count = 0;
	      detector->vad_sample_count = 0;
	      mpf_activity_detector_state_change(detector, DETECTOR_STATE_ACTIVITY_TRANSITION);
	    } else {
	      if (detector->timers_started == TRUE) {
	        detector->duration += time_base;
	        detector->noinput_duration += time_base;
	        if (detector->noinput_duration >= detector->noinput_timeout) {
	          // detected noinput
	          det_event = MPF_DETECTOR_EVENT_NOINPUT;
	          detector->noinput_duration = 0;
	        }
	      }
	    }
	  } else if (detector->state == DETECTOR_STATE_ACTIVITY_TRANSITION) {
	    detector->sample_count++;
	    if (result == TRUE) {
	      detector->vad_sample_count++;
	    }
	    if (result == TRUE || (float) detector->vad_sample_count / detector->sample_count > 0.8) {
	      detector->duration += time_base;
	      if (detector->duration >= detector->speech_timeout) {
	        // finally detected activity
	        detector->speech_duration = detector->duration;
	        det_event = MPF_DETECTOR_EVENT_ACTIVITY;
	        mpf_activity_detector_state_change(detector, DETECTOR_STATE_ACTIVITY);
	      }
	    } else {
	      // fallback to inactivity
	      detector->noinput_duration += time_base;
	      mpf_activity_detector_state_change(detector, DETECTOR_STATE_INACTIVITY);
	    }
	  } else if (detector->state == DETECTOR_STATE_ACTIVITY) {
	    if (result == TRUE) {
	      detector->duration += time_base;
	    } else {
	      // start to detect inactivity
	      detector->sample_count = 0;
	      detector->vad_sample_count = 0;
	      mpf_activity_detector_state_change(detector, DETECTOR_STATE_INACTIVITY_TRANSITION);
	    }
	  } else if (detector->state == DETECTOR_STATE_INACTIVITY_TRANSITION) {
	    detector->sample_count++;
	    if (result == TRUE) {
	      detector->vad_sample_count++;
	    }
	    if (result == TRUE && (float) detector->vad_sample_count / detector->sample_count > 0.2) {
	      // fallback to activity
	      mpf_activity_detector_state_change(detector, DETECTOR_STATE_ACTIVITY);
	    } else {
	      detector->duration += time_base;
	      if (detector->duration >= detector->silence_timeout) {
	        // detected inactivity
	        det_event = MPF_DETECTOR_EVENT_INACTIVITY;
	        detector->work = FALSE;
	        mpf_activity_detector_state_change(detector, DETECTOR_STATE_INACTIVITY);
	      }
	    }
	  }

	  return det_event;

	*/
}
