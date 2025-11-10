package gowaveform

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

// Helper function to verify an image file exists and can be opened
func verifyImageFile(t *testing.T, filename string) {
	t.Helper()

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open image file %s: %v", filename, err)
	}
	defer file.Close()

	// Try to decode the image to verify it's valid
	_, _, err = image.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode image file %s: %v", filename, err)
	}
}

func TestSavePlotBasicPNG(t *testing.T) {
	tmpWav := "/tmp/test_plot_basic.wav"
	tmpPlot := "/tmp/test_plot_basic.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save the plot with default options
	err = SavePlot(waveform, tmpPlot)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotJPEG(t *testing.T) {
	tmpWav := "/tmp/test_plot_jpeg.wav"
	tmpPlot := "/tmp/test_plot_jpeg.jpg"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save as JPEG
	err = SavePlot(waveform, tmpPlot)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotWithWidth(t *testing.T) {
	tmpWav := "/tmp/test_plot_width.wav"
	tmpPlot := "/tmp/test_plot_width.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 2.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with custom width
	err = SavePlot(waveform, tmpPlot, OptionSetWidth(1200))
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)

	// Open and check dimensions
	file, err := os.Open(tmpPlot)
	if err != nil {
		t.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()

	// Width should be approximately 1200 (may vary slightly due to DPI conversion)
	if width < 1150 || width > 1250 {
		t.Errorf("Expected width around 1200, got %d", width)
	}
}

func TestSavePlotWithHeight(t *testing.T) {
	tmpWav := "/tmp/test_plot_height.wav"
	tmpPlot := "/tmp/test_plot_height.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with custom height
	err = SavePlot(waveform, tmpPlot, OptionSetHeight(600))
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)

	// Open and check dimensions
	file, err := os.Open(tmpPlot)
	if err != nil {
		t.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode PNG: %v", err)
	}

	bounds := img.Bounds()
	height := bounds.Dy()

	// Height should be approximately 600 (may vary slightly due to DPI conversion)
	if height < 550 || height > 650 {
		t.Errorf("Expected height around 600, got %d", height)
	}
}

func TestSavePlotWithColors(t *testing.T) {
	tmpWav := "/tmp/test_plot_colors.wav"
	tmpPlot := "/tmp/test_plot_colors.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with custom colors
	err = SavePlot(waveform, tmpPlot,
		OptionSetBackgroundColor("#000000"), // Black background
		OptionSetForegroundColor("#00FF00"), // Green foreground
	)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotWithoutTimestamp(t *testing.T) {
	tmpWav := "/tmp/test_plot_no_timestamp.wav"
	tmpPlot := "/tmp/test_plot_no_timestamp.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save without timestamp
	err = SavePlot(waveform, tmpPlot, OptionShowTimestamp(false))
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotWithAllOptions(t *testing.T) {
	tmpWav := "/tmp/test_plot_all_options.wav"
	tmpPlot := "/tmp/test_plot_all_options.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 2.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with all options combined
	err = SavePlot(waveform, tmpPlot,
		OptionSetWidth(1000),
		OptionSetHeight(500),
		OptionSetBackgroundColor("#1a1a1a"),
		OptionSetForegroundColor("#ff6b35"),
		OptionShowTimestamp(true),
	)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotUnsupportedFormat(t *testing.T) {
	tmpWav := "/tmp/test_plot_unsupported.wav"
	tmpPlot := "/tmp/test_plot_unsupported.bmp"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Try to save with unsupported format
	err = SavePlot(waveform, tmpPlot)
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
}

func TestHexToColor(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected color.Color
	}{
		{
			name:     "Full hex with hash",
			hex:      "#FF0000",
			expected: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name:     "Full hex without hash",
			hex:      "00FF00",
			expected: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name:     "Short hex with hash",
			hex:      "#F0F",
			expected: color.RGBA{R: 255, G: 0, B: 255, A: 255},
		},
		{
			name:     "Short hex without hash",
			hex:      "0F0",
			expected: color.RGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name:     "Black",
			hex:      "#000000",
			expected: color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			name:     "White",
			hex:      "#FFFFFF",
			expected: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexToColor(tt.hex)
			expectedRGBA := tt.expected.(color.RGBA)
			resultRGBA := result.(color.RGBA)

			if resultRGBA.R != expectedRGBA.R ||
				resultRGBA.G != expectedRGBA.G ||
				resultRGBA.B != expectedRGBA.B ||
				resultRGBA.A != expectedRGBA.A {
				t.Errorf("hexToColor(%s) = RGBA{%d, %d, %d, %d}, want RGBA{%d, %d, %d, %d}",
					tt.hex,
					resultRGBA.R, resultRGBA.G, resultRGBA.B, resultRGBA.A,
					expectedRGBA.R, expectedRGBA.G, expectedRGBA.B, expectedRGBA.A)
			}
		})
	}
}

func TestSavePlotWithRealAudio(t *testing.T) {
	const amenFile = "data/amen_170.wav"
	tmpPlot := "/tmp/test_plot_amen.png"
	defer os.Remove(tmpPlot)

	// Check if file exists, skip if not
	if _, err := os.Stat(amenFile); os.IsNotExist(err) {
		t.Skip("Skipping test: data/amen_170.wav not found")
	}

	// Load the real waveform
	waveform, err := LoadWaveform(amenFile)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save the plot
	err = SavePlot(waveform, tmpPlot,
		OptionSetWidth(1200),
		OptionSetHeight(400),
		OptionSetBackgroundColor("#FFFFFF"),
		OptionSetForegroundColor("#0064C8"),
	)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)

	t.Logf("Successfully created plot: %s", tmpPlot)
}

func TestSavePlotJPEGWithRealAudio(t *testing.T) {
	const amenFile = "data/amen_170.wav"
	tmpPlot := "/tmp/test_plot_amen.jpeg"
	defer os.Remove(tmpPlot)

	// Check if file exists, skip if not
	if _, err := os.Stat(amenFile); os.IsNotExist(err) {
		t.Skip("Skipping test: data/amen_170.wav not found")
	}

	// Load the real waveform
	waveform, err := LoadWaveform(amenFile)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save as JPEG
	err = SavePlot(waveform, tmpPlot,
		OptionSetWidth(800),
		OptionSetHeight(300),
	)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created and can be decoded as JPEG
	file, err := os.Open(tmpPlot)
	if err != nil {
		t.Fatalf("Failed to open image file: %v", err)
	}
	defer file.Close()

	_, err = jpeg.Decode(file)
	if err != nil {
		t.Fatalf("Failed to decode JPEG: %v", err)
	}

	t.Logf("Successfully created JPEG plot: %s", tmpPlot)
}

func TestSavePlotHideYAxis(t *testing.T) {
	tmpWav := "/tmp/test_plot_hide_yaxis.wav"
	tmpPlot := "/tmp/test_plot_hide_yaxis.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with y-axis hidden
	err = SavePlot(waveform, tmpPlot, OptionHideYAxis(true))
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotWithTitle(t *testing.T) {
	tmpWav := "/tmp/test_plot_with_title.wav"
	tmpPlot := "/tmp/test_plot_with_title.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 1.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with custom title
	customTitle := "My Custom Waveform Title"
	err = SavePlot(waveform, tmpPlot, OptionSetTitle(customTitle))
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}

func TestSavePlotWithTitleAndHiddenYAxis(t *testing.T) {
	tmpWav := "/tmp/test_plot_title_hidden_yaxis.wav"
	tmpPlot := "/tmp/test_plot_title_hidden_yaxis.png"
	defer os.Remove(tmpWav)
	defer os.Remove(tmpPlot)

	// Create a test WAV file
	createTestWAV(t, tmpWav, 44100, 2.0)

	// Load the waveform
	waveform, err := LoadWaveform(tmpWav)
	if err != nil {
		t.Fatalf("LoadWaveform failed: %v", err)
	}

	// Save with custom title and hidden y-axis
	err = SavePlot(waveform, tmpPlot,
		OptionSetTitle("Audio Visualization"),
		OptionHideYAxis(true),
		OptionSetWidth(1000),
		OptionSetHeight(400),
	)
	if err != nil {
		t.Fatalf("SavePlot failed: %v", err)
	}

	// Verify the file was created
	verifyImageFile(t, tmpPlot)
}
