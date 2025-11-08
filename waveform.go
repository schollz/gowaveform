package gowaveform

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

// WaveformData represents the JSON output format compatible with audiowaveform
type WaveformData struct {
	Version         int     `json:"version"`
	Channels        int     `json:"channels"`
	SampleRate      int     `json:"sample_rate"`
	SamplesPerPixel int     `json:"samples_per_pixel"`
	Bits            int     `json:"bits"`
	Length          int     `json:"length"`
	Data            []int16 `json:"data"`
}

// WaveformOptions defines parameters for waveform generation
type WaveformOptions struct {
	Start           float64 // Start time in seconds
	End             float64 // End time in seconds (0 means end of file)
	SamplesPerPixel int     // Zoom level (samples per pixel)
}

// WAVHeader represents the WAV file header
type WAVHeader struct {
	SampleRate    uint32
	Channels      uint16
	BitsPerSample uint16
	DataSize      uint32
	DataOffset    int64
}

// ReadWAVHeader reads and parses a WAV file header
func ReadWAVHeader(r io.ReadSeeker) (*WAVHeader, error) {
	// Read RIFF header
	riffHeader := make([]byte, 12)
	if _, err := io.ReadFull(r, riffHeader); err != nil {
		return nil, fmt.Errorf("failed to read RIFF header: %w", err)
	}

	if string(riffHeader[0:4]) != "RIFF" {
		return nil, fmt.Errorf("not a valid WAV file: missing RIFF header")
	}

	if string(riffHeader[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid WAV file: missing WAVE format")
	}

	header := &WAVHeader{}

	// Read chunks until we find fmt and data
	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(r, chunkHeader); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read chunk header: %w", err)
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		switch chunkID {
		case "fmt ":
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(r, fmtData); err != nil {
				return nil, fmt.Errorf("failed to read fmt chunk: %w", err)
			}

			audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
			if audioFormat != 1 {
				return nil, fmt.Errorf("unsupported audio format: %d (only PCM is supported)", audioFormat)
			}

			header.Channels = binary.LittleEndian.Uint16(fmtData[2:4])
			header.SampleRate = binary.LittleEndian.Uint32(fmtData[4:8])
			header.BitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])

		case "data":
			header.DataSize = chunkSize
			pos, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return nil, fmt.Errorf("failed to get current position: %w", err)
			}
			header.DataOffset = pos
			// Skip data chunk for now
			if _, err := r.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("failed to skip data chunk: %w", err)
			}

		default:
			// Skip unknown chunks
			if _, err := r.Seek(int64(chunkSize), io.SeekCurrent); err != nil {
				return nil, fmt.Errorf("failed to skip chunk %s: %w", chunkID, err)
			}
		}
	}

	if header.SampleRate == 0 {
		return nil, fmt.Errorf("invalid WAV file: missing fmt chunk")
	}

	if header.DataSize == 0 {
		return nil, fmt.Errorf("invalid WAV file: missing data chunk")
	}

	return header, nil
}

// GenerateWaveformData reads a WAV file and generates waveform data
func GenerateWaveformData(filename string, opts WaveformOptions) (*WaveformData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	header, err := ReadWAVHeader(file)
	if err != nil {
		return nil, err
	}

	// Calculate sample range
	bytesPerSample := int(header.BitsPerSample) / 8
	totalSamples := int(header.DataSize) / (int(header.Channels) * bytesPerSample)

	startSample := int(opts.Start * float64(header.SampleRate))
	endSample := totalSamples
	if opts.End > 0 {
		endSample = int(opts.End * float64(header.SampleRate))
	}

	if startSample < 0 {
		startSample = 0
	}
	if endSample > totalSamples {
		endSample = totalSamples
	}
	if startSample >= endSample {
		return nil, fmt.Errorf("invalid range: start must be before end")
	}

	samplesPerPixel := opts.SamplesPerPixel
	if samplesPerPixel <= 0 {
		samplesPerPixel = 256 // Default zoom level
	}

	// Seek to start of data chunk + offset for start sample
	startOffset := header.DataOffset + int64(startSample*int(header.Channels)*bytesPerSample)
	if _, err := file.Seek(startOffset, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to start position: %w", err)
	}

	// Read and process audio data
	samplesToRead := endSample - startSample
	waveformData := &WaveformData{
		Version:         2,
		Channels:        int(header.Channels),
		SampleRate:      int(header.SampleRate),
		SamplesPerPixel: samplesPerPixel,
		Bits:            int(header.BitsPerSample),
		Length:          0,
		Data:            []int16{},
	}

	// Process samples in chunks (pixels)
	samplesRead := 0
	for samplesRead < samplesToRead {
		samplesToProcess := samplesPerPixel
		if samplesRead+samplesToProcess > samplesToRead {
			samplesToProcess = samplesToRead - samplesRead
		}

		min, max, err := readPeaks(file, header, samplesToProcess)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read peaks: %w", err)
		}

		// audiowaveform format stores min/max pairs for each channel
		// For simplicity, we'll process all channels together (mono/stereo mix)
		waveformData.Data = append(waveformData.Data, min, max)
		samplesRead += samplesToProcess
	}

	waveformData.Length = len(waveformData.Data) / 2

	return waveformData, nil
}

// readPeaks reads a chunk of samples and returns the min and max values
func readPeaks(r io.Reader, header *WAVHeader, sampleCount int) (int16, int16, error) {
	bytesPerSample := int(header.BitsPerSample) / 8
	bytesToRead := sampleCount * int(header.Channels) * bytesPerSample

	buffer := make([]byte, bytesToRead)
	n, err := io.ReadFull(r, buffer)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return 0, 0, err
	}
	if n == 0 {
		return 0, 0, io.EOF
	}

	var min, max int16 = math.MaxInt16, math.MinInt16

	// Process samples based on bit depth
	switch header.BitsPerSample {
	case 16:
		for i := 0; i < n; i += 2 {
			if i+1 >= n {
				break
			}
			sample := int16(binary.LittleEndian.Uint16(buffer[i : i+2]))
			if sample < min {
				min = sample
			}
			if sample > max {
				max = sample
			}
		}
	case 8:
		// 8-bit samples are unsigned (0-255), convert to signed
		for i := 0; i < n; i++ {
			sample := int16(buffer[i]) - 128
			sample = sample << 8 // Scale to 16-bit range
			if sample < min {
				min = sample
			}
			if sample > max {
				max = sample
			}
		}
	default:
		return 0, 0, fmt.Errorf("unsupported bit depth: %d", header.BitsPerSample)
	}

	if min == math.MaxInt16 && max == math.MinInt16 {
		// No samples were read
		min, max = 0, 0
	}

	return min, max, nil
}

// GenerateJSON generates JSON output from waveform data
func GenerateJSON(data *WaveformData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// GenerateWaveformJSON is a convenience function that generates JSON directly from a WAV file
func GenerateWaveformJSON(filename string, opts WaveformOptions) ([]byte, error) {
	data, err := GenerateWaveformData(filename, opts)
	if err != nil {
		return nil, err
	}
	return GenerateJSON(data)
}
