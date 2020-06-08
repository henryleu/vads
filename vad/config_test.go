package vad

import (
	"reflect"
	"strings"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	// type fields struct {
	// 	SpeechTimeout      int
	// 	SilenceTimeout     int
	// 	NoinputTimeout     int
	// 	NoinputTimers      bool
	// 	RecognitionTimeout int
	// 	RecognitionTimers  bool
	// 	VADLevel           VADLevel
	// 	SampleRate         int
	// 	BytesPerSample     int
	// 	FrameDuration      int
	// 	Multiple           bool
	// }

	invalidSpeechTimeout := defaultConfig
	invalidSpeechTimeout.SpeechTimeout = 0

	tests := []struct {
		name         string
		c            *Config
		wantErr      bool
		invalidField string
	}{
		// TODO: Add test cases.
		{
			name:         "Config.Validate() - invalid SpeechTimeout",
			c:            &invalidSpeechTimeout,
			wantErr:      true,
			invalidField: "SpeechTimeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.c.Validate()
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			} else if hasErr && tt.invalidField != "" {
				msg := err.Error()
				if !strings.Contains(msg, tt.invalidField) {
					t.Errorf("Config.Validate() error = %v, wantErr %v, invalidField %v", err, tt.wantErr, tt.invalidField)
				}

			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_NewDetector(t *testing.T) {
	type fields struct {
		SpeechTimeout      int
		SilenceTimeout     int
		NoinputTimeout     int
		NoinputTimers      bool
		RecognitionTimeout int
		RecognitionTimers  bool
		VADLevel           VADLevel
		SampleRate         int
		BytesPerSample     int
		FrameDuration      int
		Multiple           bool
	}
	tests := []struct {
		name   string
		fields fields
		want   *Detector
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				SpeechTimeout:      tt.fields.SpeechTimeout,
				SilenceTimeout:     tt.fields.SilenceTimeout,
				NoinputTimeout:     tt.fields.NoinputTimeout,
				NoinputTimers:      tt.fields.NoinputTimers,
				RecognitionTimeout: tt.fields.RecognitionTimeout,
				RecognitionTimers:  tt.fields.RecognitionTimers,
				VADLevel:           tt.fields.VADLevel,
				SampleRate:         tt.fields.SampleRate,
				BytesPerSample:     tt.fields.BytesPerSample,
				FrameDuration:      tt.fields.FrameDuration,
				Multiple:           tt.fields.Multiple,
			}
			if got := c.NewDetector(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.NewDetector() = %v, want %v", got, tt.want)
			}
		})
	}
}
