package cmd

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gopxl/beep/v2"
)

// WAVStream implements beep.StreamSeeker for playing WAV audio data
type WAVStream struct {
	data       []byte
	sampleRate beep.SampleRate
	position   int
	dataStart  int
	dataSize   int
}

// NewWAVStream creates a new WAV stream from raw WAV data
func NewWAVStream(wavData []byte) (*WAVStream, beep.Format, error) {
	if len(wavData) < 44 {
		return nil, beep.Format{}, fmt.Errorf("invalid WAV data: too short")
	}

	// Verify RIFF header
	if string(wavData[0:4]) != "RIFF" {
		return nil, beep.Format{}, fmt.Errorf("invalid WAV data: missing RIFF header")
	}

	// Verify WAVE format
	if string(wavData[8:12]) != "WAVE" {
		return nil, beep.Format{}, fmt.Errorf("invalid WAV data: not a WAVE file")
	}

	// Find the fmt chunk
	fmtPos := 12
	for fmtPos < len(wavData)-8 {
		chunkID := string(wavData[fmtPos : fmtPos+4])
		chunkSize := binary.LittleEndian.Uint32(wavData[fmtPos+4 : fmtPos+8])
		
		if chunkID == "fmt " {
			break
		}
		fmtPos += 8 + int(chunkSize)
	}

	if fmtPos >= len(wavData)-24 {
		return nil, beep.Format{}, fmt.Errorf("invalid WAV data: fmt chunk not found")
	}

	// Read format data
	fmtData := wavData[fmtPos+8 : fmtPos+8+16]
	audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
	numChannels := binary.LittleEndian.Uint16(fmtData[2:4])
	sampleRate := binary.LittleEndian.Uint32(fmtData[4:8])
	bitsPerSample := binary.LittleEndian.Uint16(fmtData[14:16])

	if audioFormat != 1 {
		return nil, beep.Format{}, fmt.Errorf("unsupported audio format: %d (only PCM supported)", audioFormat)
	}

	if bitsPerSample != 16 {
		return nil, beep.Format{}, fmt.Errorf("unsupported bits per sample: %d (only 16-bit supported)", bitsPerSample)
	}

	// Find the data chunk
	dataPos := fmtPos + 8 + int(binary.LittleEndian.Uint32(wavData[fmtPos+4:fmtPos+8]))
	for dataPos < len(wavData)-8 {
		chunkID := string(wavData[dataPos : dataPos+4])
		chunkSize := binary.LittleEndian.Uint32(wavData[dataPos+4 : dataPos+8])
		
		if chunkID == "data" {
			dataStart := dataPos + 8
			dataSize := int(chunkSize)
			
			format := beep.Format{
				SampleRate:  beep.SampleRate(sampleRate),
				NumChannels: int(numChannels),
				Precision:   2, // 16-bit
			}

			return &WAVStream{
				data:       wavData,
				sampleRate: beep.SampleRate(sampleRate),
				position:   0,
				dataStart:  dataStart,
				dataSize:   dataSize,
			}, format, nil
		}
		dataPos += 8 + int(chunkSize)
	}

	return nil, beep.Format{}, fmt.Errorf("invalid WAV data: data chunk not found")
}

func (s *WAVStream) Stream(samples [][2]float64) (n int, ok bool) {
	bytesPerSample := 2 // 16-bit
	currentPos := s.dataStart + s.position

	if s.position >= s.dataSize {
		return 0, false
	}

	for i := range samples {
		if s.position+bytesPerSample > s.dataSize {
			return i, true
		}

		// Convert 16-bit little-endian PCM to float64
		sample16 := int16(s.data[currentPos]) | int16(s.data[currentPos+1])<<8
		sampleFloat := float64(sample16) / 32768.0

		// Mono to stereo (assuming mono input)
		samples[i][0] = sampleFloat
		samples[i][1] = sampleFloat

		s.position += bytesPerSample
		currentPos += bytesPerSample
	}

	return len(samples), true
}

func (s *WAVStream) Err() error {
	return nil
}

func (s *WAVStream) Len() int {
	return s.dataSize / 2 // 16-bit samples
}

func (s *WAVStream) Position() int {
	return s.position / 2
}

func (s *WAVStream) Seek(p int) error {
	s.position = p * 2
	if s.position < 0 {
		s.position = 0
	}
	if s.position > s.dataSize {
		s.position = s.dataSize
	}
	return nil
}

// DecodeWAV decodes WAV data and returns a StreamSeeker
func DecodeWAV(r io.Reader) (beep.StreamSeeker, beep.Format, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, beep.Format{}, err
	}
	
	stream, format, err := NewWAVStream(data)
	if err != nil {
		return nil, beep.Format{}, err
	}
	
	return stream, format, nil
}