package gowaveform

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Waveform represents a loaded WAV file with its audio data
type Waveform struct {
	SampleRate      int
	Channels        int
	BitsPerSample   int
	audioData       []int16 // All audio samples in int16 format (interleaved for multi-channel)
	totalSamples    int     // Total number of frames (not individual channel samples)
}

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
	SamplesPerPixel int     // Zoom level (samples per pixel). Ignored if Width is specified.
	Width           int     // Target width in pixels. If specified, SamplesPerPixel is calculated automatically.
}

// WAVHeader represents the WAV file header
type WAVHeader struct {
	SampleRate    uint32
	Channels      uint16
	BitsPerSample uint16
	DataSize      uint32
	DataOffset    int64
}

// LoadWaveform loads a WAV file into memory for generating multiple views
func LoadWaveform(filename string) (*Waveform, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create decoder
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	// Check for PCM format (or extensible format which may contain PCM)
	// Format 1 = PCM, Format 65534 = WAVE_FORMAT_EXTENSIBLE
	if decoder.WavAudioFormat != 1 && decoder.WavAudioFormat != 65534 {
		return nil, fmt.Errorf("unsupported audio format: %d (only PCM is supported)", decoder.WavAudioFormat)
	}

	// Forward to PCM chunk
	if err := decoder.FwdToPCM(); err != nil {
		return nil, fmt.Errorf("failed to read PCM chunk: %w", err)
	}

	// Calculate total samples (frames)
	bytesPerSample := int(decoder.BitDepth) / 8
	totalSamples := decoder.PCMSize / (int(decoder.NumChans) * bytesPerSample)

	// Read all audio data into memory
	audioData := make([]int16, 0, totalSamples*int(decoder.NumChans))
	
	// Read in chunks
	bufferSize := 4096 * int(decoder.NumChans)
	intBuf := &audio.IntBuffer{
		Data:   make([]int, bufferSize),
		Format: decoder.Format(),
	}

	bitDepth := int(decoder.BitDepth)
	for {
		n, err := decoder.PCMBuffer(intBuf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read audio data: %w", err)
		}
		if n == 0 {
			break
		}

		// Convert to int16 and append
		for i := 0; i < n; i++ {
			var sample int16
			switch bitDepth {
			case 16:
				sample = int16(intBuf.Data[i])
			case 8:
				// 8-bit samples are in the range 0-255, convert to signed 16-bit
				sample = int16((intBuf.Data[i] - 128) << 8)
			case 24:
				// 24-bit samples, scale to 16-bit
				sample = int16(intBuf.Data[i] >> 8)
			case 32:
				// 32-bit samples, scale to 16-bit
				sample = int16(intBuf.Data[i] >> 16)
			default:
				sample = int16(intBuf.Data[i])
			}
			audioData = append(audioData, sample)
		}

		if err == io.EOF {
			break
		}
	}

	waveform := &Waveform{
		SampleRate:    int(decoder.SampleRate),
		Channels:      int(decoder.NumChans),
		BitsPerSample: int(decoder.BitDepth),
		audioData:     audioData,
		totalSamples:  totalSamples,
	}

	return waveform, nil
}

// Duration returns the total duration of the audio in seconds
func (w *Waveform) Duration() float64 {
	if w.SampleRate == 0 {
		return 0
	}
	return float64(w.totalSamples) / float64(w.SampleRate)
}

// GenerateView generates a waveform view from the loaded audio data
func (w *Waveform) GenerateView(opts WaveformOptions) (*WaveformData, error) {
	startSample := int(opts.Start * float64(w.SampleRate))
	endSample := w.totalSamples
	if opts.End > 0 {
		endSample = int(opts.End * float64(w.SampleRate))
	}

	if startSample < 0 {
		startSample = 0
	}
	if endSample > w.totalSamples {
		endSample = w.totalSamples
	}
	if startSample >= endSample {
		return nil, fmt.Errorf("invalid range: start must be before end")
	}

	// Calculate samples per pixel based on width or use the specified value
	samplesPerPixel := opts.SamplesPerPixel
	if opts.Width > 0 {
		// Calculate zoom level to fit the requested range into the specified width
		samplesToRead := endSample - startSample
		samplesPerPixel = samplesToRead / opts.Width
		if samplesPerPixel <= 0 {
			samplesPerPixel = 1 // Minimum zoom level
		}
	} else if samplesPerPixel <= 0 {
		samplesPerPixel = 256 // Default zoom level
	}

	// Initialize waveform data
	waveformData := &WaveformData{
		Version:         2,
		Channels:        w.Channels,
		SampleRate:      w.SampleRate,
		SamplesPerPixel: samplesPerPixel,
		Bits:            w.BitsPerSample,
		Length:          0,
		Data:            []int16{},
	}

	// Process the range
	samplesToRead := endSample - startSample
	samplesRead := 0

	for samplesRead < samplesToRead {
		samplesToProcess := samplesPerPixel
		if samplesRead+samplesToProcess > samplesToRead {
			samplesToProcess = samplesToRead - samplesRead
		}

		// Calculate min/max from audio data
		currentSample := startSample + samplesRead
		min, max := w.getPeaksFromRange(currentSample, samplesToProcess)

		waveformData.Data = append(waveformData.Data, min, max)
		samplesRead += samplesToProcess
	}

	waveformData.Length = len(waveformData.Data) / 2

	return waveformData, nil
}

// getPeaksFromRange calculates min and max peaks from a range of samples in the audio data
func (w *Waveform) getPeaksFromRange(startSample, sampleCount int) (int16, int16) {
	var min, max int16 = math.MaxInt16, math.MinInt16

	endSample := startSample + sampleCount
	if endSample > w.totalSamples {
		endSample = w.totalSamples
	}

	// Calculate the starting and ending indices in the interleaved audio data
	startIdx := startSample * w.Channels
	endIdx := endSample * w.Channels

	if startIdx >= len(w.audioData) {
		return 0, 0
	}
	if endIdx > len(w.audioData) {
		endIdx = len(w.audioData)
	}

	// Process all samples in the range (all channels)
	for i := startIdx; i < endIdx; i++ {
		sample := w.audioData[i]
		if sample < min {
			min = sample
		}
		if sample > max {
			max = sample
		}
	}

	if min == math.MaxInt16 && max == math.MinInt16 {
		min, max = 0, 0
	}

	return min, max
}

// ReadWAVHeader reads and parses a WAV file header using go-audio/wav
func ReadWAVHeader(r io.ReadSeeker) (*WAVHeader, error) {
	// Create decoder
	decoder := wav.NewDecoder(r)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("not a valid WAV file")
	}

	// Check for PCM format (or extensible format which may contain PCM)
	// Format 1 = PCM, Format 65534 = WAVE_FORMAT_EXTENSIBLE
	if decoder.WavAudioFormat != 1 && decoder.WavAudioFormat != 65534 {
		return nil, fmt.Errorf("unsupported audio format: %d (only PCM is supported)", decoder.WavAudioFormat)
	}

	// Forward to PCM chunk to get size information
	if err := decoder.FwdToPCM(); err != nil {
		return nil, fmt.Errorf("failed to read PCM chunk: %w", err)
	}

	header := &WAVHeader{
		SampleRate:    decoder.SampleRate,
		Channels:      decoder.NumChans,
		BitsPerSample: decoder.BitDepth,
		DataSize:      uint32(decoder.PCMSize),
		DataOffset:    0, // The decoder handles positioning internally
	}

	return header, nil
}

// GenerateWaveformData reads a WAV file and generates waveform data
func GenerateWaveformData(filename string, opts WaveformOptions) (*WaveformData, error) {
	// Use LoadWaveform + GenerateView approach to avoid decoder.Seek issues
	waveform, err := LoadWaveform(filename)
	if err != nil {
		return nil, err
	}

	return waveform.GenerateView(opts)
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
