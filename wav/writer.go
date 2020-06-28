package wave

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// WriterParam defines the parameters for creating a new writer
type WriterParam struct {
	Out           io.WriteCloser
	Channel       int
	SampleRate    int
	BitsPerSample int
}

// Debug prints debug info
func (wp WriterParam) Debug() {
	log.Printf("Channel:\t%v\n", wp.Channel)
	log.Printf("SampleRate:\t%v\n", wp.SampleRate)
	log.Printf("BitsPerSample:\t%v\n", wp.BitsPerSample)
}

// Writer defines a writer to write data to a wave file
type Writer struct {
	out            io.WriteCloser // the interface of the created file to write
	writtenSamples int            // the number of samples written

	riffChunk *RiffChunk
	fmtChunk  *FmtChunk
	dataChunk *DataWriterChunk
}

// NewWriter creates and writes wave data to a wave file.
func NewWriter(param WriterParam) (*Writer, error) {
	w := &Writer{}
	w.out = param.Out

	blockSize := uint16(param.BitsPerSample*param.Channel) / 8
	samplesPerSec := uint32(int(blockSize) * param.SampleRate)
	//	fmt.Println(blockSize, param.SampleRate, samplesPerSec)

	// riff chunk
	w.riffChunk = &RiffChunk{
		ID:         []byte(riffChunkToken),
		FormatType: []byte(waveFormatType),
	}
	// fmt chunk
	w.fmtChunk = &FmtChunk{
		ID:   []byte(fmtChunkToken),
		Size: uint32(fmtChunkSize),
	}
	w.fmtChunk.Data = &WavFmtChunkData{
		WaveFormatType: uint16(1), // PCM
		Channel:        uint16(param.Channel),
		SamplesPerSec:  uint32(param.SampleRate),
		BytesPerSec:    samplesPerSec,
		BlockSize:      uint16(blockSize),
		BitsPerSamples: uint16(param.BitsPerSample),
	}
	// data chunk
	w.dataChunk = &DataWriterChunk{
		ID:   []byte(dataChunkToken),
		Data: bytes.NewBuffer([]byte{}),
	}

	return w, nil
}

// WriteSample8 writes uint8 slice samples to the writer
func (w *Writer) WriteSample8(samples []uint8) (int, error) {
	buf := new(bytes.Buffer)

	for i := 0; i < len(samples); i++ {
		err := binary.Write(buf, binary.LittleEndian, samples[i])
		if err != nil {
			return 0, err
		}
	}
	n, err := w.Write(buf.Bytes())
	return n, err
}

// WriteSample16 writes int16 slice samples to the writer
func (w *Writer) WriteSample16(samples []int16) (int, error) {
	buf := new(bytes.Buffer)

	for i := 0; i < len(samples); i++ {
		err := binary.Write(buf, binary.LittleEndian, samples[i])
		if err != nil {
			return 0, err
		}
	}
	n, err := w.Write(buf.Bytes())
	return n, err
}

// Write writes byte slice as samples to the writer
func (w *Writer) Write(p []byte) (int, error) {
	blockSize := int(w.fmtChunk.Data.BlockSize)
	if len(p) < blockSize {
		return 0, fmt.Errorf("writing data need at least %d bytes", blockSize)
	}

	// The number of write bytes is a multiple of BlockSize
	if len(p)%blockSize != 0 {
		return 0, fmt.Errorf("writing data must be a multiple of %d bytes", blockSize)
	}
	num := len(p) / blockSize

	n, err := w.dataChunk.Data.Write(p)

	if err == nil {
		w.writtenSamples += num
	}
	return n, err
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) Write(order binary.ByteOrder, data interface{}) {
	if ew.err != nil {
		return
	}
	ew.err = binary.Write(ew.w, order, data)
}

// Close closes the created file
func (w *Writer) Close() error {
	data := w.dataChunk.Data.Bytes()
	dataSize := uint32(len(data))
	w.riffChunk.Size = uint32(len(w.riffChunk.ID)) + (8 + w.fmtChunk.Size) + (8 + dataSize)
	w.dataChunk.Size = dataSize

	ew := &errWriter{w: w.out}
	// riff chunk
	ew.Write(binary.BigEndian, w.riffChunk.ID)
	ew.Write(binary.LittleEndian, w.riffChunk.Size)
	ew.Write(binary.BigEndian, w.riffChunk.FormatType)

	// fmt chunk
	ew.Write(binary.BigEndian, w.fmtChunk.ID)
	ew.Write(binary.LittleEndian, w.fmtChunk.Size)
	ew.Write(binary.LittleEndian, w.fmtChunk.Data)

	//data chunk
	ew.Write(binary.BigEndian, w.dataChunk.ID)
	ew.Write(binary.LittleEndian, w.dataChunk.Size)

	if ew.err != nil {
		return ew.err
	}

	_, err := w.out.Write(data)
	if err != nil {
		return err
	}

	err = w.out.Close()
	if err != nil {
		return err
	}

	return nil
}
