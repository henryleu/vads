package wave

import (
	"bytes"
	"io"
)

const (
	maxFileSize             = 2 << 31
	riffChunkSize           = 12
	listChunkOffset         = 36
	riffChunkSizeBaseOffset = 36 // RIFFChunk(12byte) + fmtChunk(24byte) = 36byte
	fmtChunkSize            = 16
)

var (
	riffChunkToken = "RIFF"
	waveFormatType = "WAVE"
	fmtChunkToken  = "fmt "
	listChunkToken = "LIST"
	dataChunkToken = "data"
)

// RiffChunk has 12 bytes
type RiffChunk struct {
	ID         []byte // 'RIFF'
	Size       uint32 // 36bytes + data_chunk_size or whole_file_size - 'RIFF'+ChunkSize (8byte)
	FormatType []byte // 'WAVE'
}

// FmtChunk is with 8 + 16 = 24 bytes
type FmtChunk struct {
	ID   []byte // 'fmt '
	Size uint32 // 16
	Data *WavFmtChunkData
}

// WavFmtChunkData is with 16 bytes
type WavFmtChunkData struct {
	WaveFormatType uint16 // PCM は 1
	Channel        uint16 // monoral or stereo
	SamplesPerSec  uint32 // Sampling frequency: 8000, 16000 or 44100
	BytesPerSec    uint32 // the number of bytes required per second
	BlockSize      uint16 // Quantization accuracy * Number of channels
	BitsPerSamples uint16
}

// DataReader defines interface for DataReader
type DataReader interface {
	io.Reader
	io.ReaderAt
}

// DataReaderChunk defines a data chunk with reader
type DataReaderChunk struct {
	ID   []byte     // 'data'
	Size uint32     // sound data length * channels
	Data DataReader // Actual data
}

// DataWriterChunk defines a data chunk with a writer
type DataWriterChunk struct {
	ID   []byte
	Size uint32
	Data *bytes.Buffer
}
