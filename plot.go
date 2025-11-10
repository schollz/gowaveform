package gowaveform

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// PlotConfig holds the configuration for plotting a waveform
type PlotConfig struct {
	width           int
	height          int
	backgroundColor color.Color
	foregroundColor color.Color
	showTimestamp   bool
}

// Option is the type all plot options need to adhere to
type Option func(*PlotConfig)

// OptionSetWidth sets the width of the plot in pixels
func OptionSetWidth(width int) Option {
	return func(c *PlotConfig) {
		c.width = width
	}
}

// OptionSetHeight sets the height of the plot in pixels
func OptionSetHeight(height int) Option {
	return func(c *PlotConfig) {
		c.height = height
	}
}

// OptionSetBackgroundColor sets the background color using a hex color code
func OptionSetBackgroundColor(hexColor string) Option {
	return func(c *PlotConfig) {
		c.backgroundColor = hexToColor(hexColor)
	}
}

// OptionSetForegroundColor sets the foreground color (waveform color) using a hex color code
func OptionSetForegroundColor(hexColor string) Option {
	return func(c *PlotConfig) {
		c.foregroundColor = hexToColor(hexColor)
	}
}

// OptionShowTimestamp enables or disables the timestamp display below the waveform
func OptionShowTimestamp(show bool) Option {
	return func(c *PlotConfig) {
		c.showTimestamp = show
	}
}

// hexToColor converts a hex color string to color.Color
// Supports formats: #RGB, #RRGGBB, RGB, RRGGBB
func hexToColor(hex string) color.Color {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	// Default to black if invalid
	if len(hex) != 3 && len(hex) != 6 {
		return color.Black
	}

	// Expand 3-digit hex to 6-digit
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// SavePlot saves the waveform visualization to an image file
// The file format (PNG or JPEG) is determined by the filename extension
func SavePlot(w *Waveform, filename string, opts ...Option) error {
	// Default configuration
	config := PlotConfig{
		width:           800,
		height:          400,
		backgroundColor: color.White,
		foregroundColor: color.RGBA{R: 0, G: 100, B: 200, A: 255}, // Blue
		showTimestamp:   true,
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Generate waveform data
	waveformData, err := w.GenerateView(WaveformOptions{
		Start: 0,
		End:   0, // Use full duration
		Width: config.width,
	})
	if err != nil {
		return fmt.Errorf("failed to generate waveform view: %w", err)
	}

	// Create a new plot
	p := plot.New()

	// Set background color
	p.BackgroundColor = config.backgroundColor

	// Set title and labels
	p.Title.Text = "Waveform"
	if config.showTimestamp {
		p.X.Label.Text = "Time (seconds)"
	}
	p.Y.Label.Text = "Amplitude"

	// Hide labels if timestamp is disabled
	if !config.showTimestamp {
		p.X.Label.Text = ""
		p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{})
	}

	// Create XY points from waveform data
	// We'll use a polygon to create the filled waveform visualization
	points := make(plotter.XYs, 0, len(waveformData.Data))
	
	duration := w.Duration()
	samplesPerPixel := waveformData.SamplesPerPixel
	
	// Create the waveform shape by plotting min/max pairs
	for i := 0; i < waveformData.Length; i++ {
		maxVal := waveformData.Data[i*2+1]
		
		// Calculate time position for this pixel
		samplePos := float64(i * samplesPerPixel)
		timePos := samplePos / float64(w.SampleRate)
		
		// Normalize amplitude to -1.0 to 1.0 range
		maxNorm := float64(maxVal) / 32768.0
		
		// Add points for the waveform
		points = append(points, plotter.XY{X: timePos, Y: maxNorm})
	}
	
	// Add points in reverse for the bottom of the waveform
	for i := waveformData.Length - 1; i >= 0; i-- {
		minVal := waveformData.Data[i*2]
		
		samplePos := float64(i * samplesPerPixel)
		timePos := samplePos / float64(w.SampleRate)
		minNormVal := float64(minVal) / 32768.0
		
		points = append(points, plotter.XY{X: timePos, Y: minNormVal})
	}

	// Create a polygon for filled waveform
	poly, err := plotter.NewPolygon(points)
	if err != nil {
		return fmt.Errorf("failed to create polygon: %w", err)
	}
	poly.Color = config.foregroundColor
	poly.LineStyle.Width = vg.Points(0) // No outline

	p.Add(poly)

	// Set X axis range to match duration
	p.X.Min = 0
	p.X.Max = duration

	// Set Y axis range
	p.Y.Min = -1.0
	p.Y.Max = 1.0

	// Determine file format from extension
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Convert pixels to vg.Length (assuming 96 DPI)
	width := vg.Length(config.width) * vg.Inch / 96
	height := vg.Length(config.height) * vg.Inch / 96

	// Save the plot
	switch ext {
	case ".png":
		if err := p.Save(width, height, filename); err != nil {
			return fmt.Errorf("failed to save PNG: %w", err)
		}
	case ".jpg", ".jpeg":
		if err := p.Save(width, height, filename); err != nil {
			return fmt.Errorf("failed to save JPEG: %w", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s (supported: .png, .jpg, .jpeg)", ext)
	}

	return nil
}
