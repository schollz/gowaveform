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
	hideYAxis       bool
	hideXAxis       bool
	title           string
	start           float64 // Start time in seconds (0 = beginning)
	end             float64 // End time in seconds (0 = use full duration)
	resolution      float64 // Resolution multiplier (1.0 = full resolution, 0.5 = half resolution)
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

// OptionHideYAxis enables or disables the y-axis display
func OptionHideYAxis(hide bool) Option {
	return func(c *PlotConfig) {
		c.hideYAxis = hide
	}
}

// OptionHideXAxis enables or disables the x-axis display
func OptionHideXAxis(hide bool) Option {
	return func(c *PlotConfig) {
		c.hideXAxis = hide
	}
}

// OptionSetTitle sets the title for the plot
func OptionSetTitle(title string) Option {
	return func(c *PlotConfig) {
		c.title = title
	}
}

// OptionSetStart sets the start time in seconds for the waveform view
func OptionSetStart(start float64) Option {
	return func(c *PlotConfig) {
		c.start = start
	}
}

// OptionSetEnd sets the end time in seconds for the waveform view
func OptionSetEnd(end float64) Option {
	return func(c *PlotConfig) {
		c.end = end
	}
}

// OptionSetZoom sets the duration (in seconds) to display, centered around the midpoint
// If start is set, zoom from start; if end is set, zoom backwards from end
func OptionSetZoom(duration float64) Option {
	return func(c *PlotConfig) {
		// This will be handled specially in SavePlot based on start/end values
		// For now, we'll store it as the end value and process it later
		if c.start > 0 {
			c.end = c.start + duration
		} else {
			// Will be calculated in SavePlot when we know total duration
			c.end = -duration // Negative indicates zoom duration
		}
	}
}

// OptionSetResolution sets the resolution multiplier for waveform generation
// 1.0 = full resolution (1 pixel per width unit)
// 0.5 = half resolution (generate with half the width)
// 2.0 = double resolution (generate with double the width)
func OptionSetResolution(resolution float64) Option {
	return func(c *PlotConfig) {
		if resolution > 0 {
			c.resolution = resolution
		}
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
		hideYAxis:       false,
		hideXAxis:       false,
		title:           "",
		start:           0,
		end:             0,
		resolution:      1.0,
	}

	// Apply options
	for _, opt := range opts {
		opt(&config)
	}

	// Get total duration
	totalDuration := w.Duration()

	// Handle zoom (negative end indicates zoom duration was set)
	if config.end < 0 {
		zoomDuration := -config.end
		if config.start > 0 {
			// Zoom from start
			config.end = config.start + zoomDuration
		} else {
			// Center zoom around midpoint
			center := totalDuration / 2.0
			config.start = center - zoomDuration/2.0
			config.end = center + zoomDuration/2.0
		}
	}

	// Clamp start and end to valid range
	if config.start < 0 {
		config.start = 0
	}
	if config.end > totalDuration || config.end == 0 {
		config.end = totalDuration
	}
	if config.start >= config.end {
		config.start = 0
		config.end = totalDuration
	}

	// Calculate effective width based on resolution
	effectiveWidth := int(float64(config.width) * config.resolution)
	if effectiveWidth < 1 {
		effectiveWidth = 1
	}

	// Generate waveform data
	waveformData, err := w.GenerateView(WaveformOptions{
		Start: config.start,
		End:   config.end,
		Width: effectiveWidth,
	})
	if err != nil {
		return fmt.Errorf("failed to generate waveform view: %w", err)
	}

	// Create a new plot
	p := plot.New()

	// Set background color
	p.BackgroundColor = config.backgroundColor

	// Set title
	p.Title.Text = config.title
	
	// Set labels
	if config.showTimestamp {
		p.X.Label.Text = "Time (seconds)"
	}
	
	if !config.hideYAxis {
		p.Y.Label.Text = "Amplitude"
	}

	// Hide labels if timestamp is disabled
	if !config.showTimestamp {
		p.X.Label.Text = ""
		p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{})
		p.X.Tick.LineStyle.Width = 0
		p.X.LineStyle.Width = 0
	}

	// Hide x-axis if requested
	if config.hideXAxis {
		p.X.Label.Text = ""
		p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{})
		p.X.Tick.LineStyle.Width = 0
		p.X.LineStyle.Width = 0
	}

	// Hide y-axis if requested
	if config.hideYAxis {
		p.Y.Label.Text = ""
		p.Y.Tick.Marker = plot.ConstantTicks([]plot.Tick{})
		p.Y.Tick.LineStyle.Width = 0
		p.Y.LineStyle.Width = 0
	}

	// Create XY points from waveform data
	// We'll use a polygon to create the filled waveform visualization
	points := make(plotter.XYs, 0, len(waveformData.Data))

	samplesPerPixel := waveformData.SamplesPerPixel

	// Create the waveform shape by plotting min/max pairs
	for i := 0; i < waveformData.Length; i++ {
		maxVal := waveformData.Data[i*2+1]

		// Calculate time position for this pixel relative to the view start
		samplePos := float64(i * samplesPerPixel)
		timePos := config.start + (samplePos / float64(w.SampleRate))

		// Normalize amplitude to -1.0 to 1.0 range
		maxNorm := float64(maxVal) / 32768.0

		// Add points for the waveform
		points = append(points, plotter.XY{X: timePos, Y: maxNorm})
	}

	// Add points in reverse for the bottom of the waveform
	for i := waveformData.Length - 1; i >= 0; i-- {
		minVal := waveformData.Data[i*2]

		samplePos := float64(i * samplesPerPixel)
		timePos := config.start + (samplePos / float64(w.SampleRate))
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

	// Set X axis range to match the view
	p.X.Min = config.start
	p.X.Max = config.end

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
