package gowaveform

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// createTestWAV creates a simple test WAV file
func createTestWAV(t *testing.T, filename string, sampleRate uint32, duration float64) {
	t.Helper()

	numSamples := int(float64(sampleRate) * duration)
	dataSize := uint32(numSamples * 2) // 16-bit mono

	buf := new(bytes.Buffer)

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+dataSize)) // File size - 8
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))  // PCM format
	binary.Write(buf, binary.LittleEndian, uint16(1))  // Mono
	binary.Write(buf, binary.LittleEndian, sampleRate)
	binary.Write(buf, binary.LittleEndian, sampleRate*2) // Byte rate
	binary.Write(buf, binary.LittleEndian, uint16(2))    // Block align
	binary.Write(buf, binary.LittleEndian, uint16(16))   // Bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, dataSize)

	// Generate sine wave test data
	for i := 0; i < numSamples; i++ {
		// Simple sine wave pattern
		sample := int16(10000 * (i % 100) / 100)
		binary.Write(buf, binary.LittleEndian, sample)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to create test WAV file: %v", err)
	}
}

func TestReadWAVHeader(t *testing.T) {
	tmpFile := "/tmp/test_header.wav"
	defer os.Remove(tmpFile)

	createTestWAV(t, tmpFile, 44100, 1.0)

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	header, err := ReadWAVHeader(file)
	if err != nil {
		t.Fatalf("ReadWAVHeader failed: %v", err)
	}

	if header.SampleRate != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", header.SampleRate)
	}

	if header.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", header.Channels)
	}

	if header.BitsPerSample != 16 {
		t.Errorf("Expected 16 bits per sample, got %d", header.BitsPerSample)
	}
}

func TestGenerateWaveformData(t *testing.T) {
	tmpFile := "/tmp/test_waveform.wav"
	defer os.Remove(tmpFile)

	createTestWAV(t, tmpFile, 44100, 1.0)

	opts := WaveformOptions{
		Start:           0,
		End:             0, // End of file
		SamplesPerPixel: 256,
	}

	data, err := GenerateWaveformData(tmpFile, opts)
	if err != nil {
		t.Fatalf("GenerateWaveformData failed: %v", err)
	}

	if data.Version != 2 {
		t.Errorf("Expected version 2, got %d", data.Version)
	}

	if data.SampleRate != 44100 {
		t.Errorf("Expected sample rate 44100, got %d", data.SampleRate)
	}

	if data.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", data.Channels)
	}

	if data.SamplesPerPixel != 256 {
		t.Errorf("Expected 256 samples per pixel, got %d", data.SamplesPerPixel)
	}

	if len(data.Data) == 0 {
		t.Error("Expected non-empty data array")
	}

	// Data should be in min/max pairs
	if len(data.Data)%2 != 0 {
		t.Error("Data length should be even (min/max pairs)")
	}

	if data.Length != len(data.Data)/2 {
		t.Errorf("Expected length %d, got %d", len(data.Data)/2, data.Length)
	}
}

func TestGenerateWaveformDataWithRange(t *testing.T) {
	tmpFile := "/tmp/test_range.wav"
	defer os.Remove(tmpFile)

	createTestWAV(t, tmpFile, 44100, 2.0)

	opts := WaveformOptions{
		Start:           0.5, // Start at 0.5 seconds
		End:             1.5, // End at 1.5 seconds
		SamplesPerPixel: 128,
	}

	data, err := GenerateWaveformData(tmpFile, opts)
	if err != nil {
		t.Fatalf("GenerateWaveformData failed: %v", err)
	}

	// Should have approximately 1 second of data
	// At 44100 Hz with 128 samples per pixel: 44100 / 128 ≈ 344 pixels
	expectedPixels := (1.0 * 44100) / 128
	actualPixels := float64(data.Length)

	// Allow some tolerance
	if actualPixels < expectedPixels*0.9 || actualPixels > expectedPixels*1.1 {
		t.Errorf("Expected approximately %.0f pixels, got %.0f", expectedPixels, actualPixels)
	}
}

func TestGenerateWaveformJSON(t *testing.T) {
	tmpFile := "/tmp/test_json.wav"
	defer os.Remove(tmpFile)

	createTestWAV(t, tmpFile, 44100, 0.5)

	opts := WaveformOptions{
		Start:           0,
		End:             0,
		SamplesPerPixel: 256,
	}

	jsonData, err := GenerateWaveformJSON(tmpFile, opts)
	if err != nil {
		t.Fatalf("GenerateWaveformJSON failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Check that it's valid JSON by checking for common fields
	jsonStr := string(jsonData)
	if !bytes.Contains(jsonData, []byte("\"version\"")) {
		t.Error("JSON missing 'version' field")
	}
	if !bytes.Contains(jsonData, []byte("\"sample_rate\"")) {
		t.Error("JSON missing 'sample_rate' field")
	}
	if !bytes.Contains(jsonData, []byte("\"data\"")) {
		t.Error("JSON missing 'data' field")
	}

	t.Logf("Generated JSON sample:\n%s", jsonStr[:min(len(jsonStr), 500)])
}

func TestReadPeaks(t *testing.T) {
	// Create a simple buffer with known values
	header := &WAVHeader{
		SampleRate:    44100,
		Channels:      1,
		BitsPerSample: 16,
	}

	// Create test data: 100, -100, 200, -200
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int16(100))
	binary.Write(buf, binary.LittleEndian, int16(-100))
	binary.Write(buf, binary.LittleEndian, int16(200))
	binary.Write(buf, binary.LittleEndian, int16(-200))

	min, max, err := readPeaks(buf, header, 4)
	if err != nil {
		t.Fatalf("readPeaks failed: %v", err)
	}

	if min != -200 {
		t.Errorf("Expected min -200, got %d", min)
	}

	if max != 200 {
		t.Errorf("Expected max 200, got %d", max)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestInvalidWAVFile(t *testing.T) {
	tmpFile := "/tmp/test_invalid.wav"
	defer os.Remove(tmpFile)

	// Create an invalid WAV file
	if err := os.WriteFile(tmpFile, []byte("not a wav file"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	file, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	_, err = ReadWAVHeader(file)
	if err == nil {
		t.Error("Expected error for invalid WAV file, got nil")
	}
}

func TestNonExistentFile(t *testing.T) {
	opts := WaveformOptions{
		Start:           0,
		End:             0,
		SamplesPerPixel: 256,
	}

	_, err := GenerateWaveformData("/tmp/nonexistent_file.wav", opts)
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// Benchmark the waveform generation
func BenchmarkGenerateWaveformData(b *testing.B) {
	tmpFile := "/tmp/bench_waveform.wav"
	defer os.Remove(tmpFile)

	// Create a 5 second test file
	createTestWAV(&testing.T{}, tmpFile, 44100, 5.0)

	opts := WaveformOptions{
		Start:           0,
		End:             0,
		SamplesPerPixel: 256,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateWaveformData(tmpFile, opts)
		if err != nil {
			b.Fatalf("GenerateWaveformData failed: %v", err)
		}
	}
}
