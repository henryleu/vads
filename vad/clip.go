package vad

// Clip defines voice clip for processing and persisting
type Clip struct {
	// SampleRate defines the number of samples per second, aka. sample rate.
	SampleRate int

	// BytesPerSample defines bytes per sample (sample depth) for linear pcm
	BytesPerSample int

	// Time defines the starting time of the voice clip in the whole voice
	// stream in milliseconds.
	Start int

	// Time defines the time span of the voice clip in milliseconds.
	Time int

	// Data is the chunk data of the voice clip as the specific sample rate and depth
	Data []byte
}
