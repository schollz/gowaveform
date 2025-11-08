package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/schollz/gowaveform"
)

type model struct {
	wavFile     string
	waveform    *gowaveform.Waveform
	currentView *gowaveform.WaveformData
	width       int
	height      int

	// Navigation state
	start         float64 // Start time in seconds
	end           float64 // End time in seconds
	totalDuration float64 // Total duration of the audio file

	// Error handling
	err error
}

func initialModel(wavFile string) model {
	return model{
		wavFile: wavFile,
		start:   0.0,
		end:     0.0, // Will be set to total duration
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Load waveform on first window size message if not already loaded
		if m.waveform == nil {
			wf, err := gowaveform.LoadWaveform(m.wavFile)
			if err != nil {
				m.err = fmt.Errorf("failed to load waveform: %w", err)
				return m, tea.Quit
			}
			m.waveform = wf

			// Calculate total duration
			m.totalDuration = wf.Duration()
			m.end = m.totalDuration
		}

		// Generate view with current width
		if m.waveform != nil {
			view, err := m.waveform.GenerateView(gowaveform.WaveformOptions{
				Start: m.start,
				End:   m.end,
				Width: m.width,
			})
			if err != nil {
				m.err = fmt.Errorf("failed to generate view: %w", err)
				return m, tea.Quit
			}
			m.currentView = view
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "left":
			// Jog left - shift the view earlier in time
			duration := m.end - m.start
			step := duration * 0.005 // Move 0.5% of current view

			m.start -= step
			m.end -= step

			// Clamp to valid range
			if m.start < 0 {
				m.start = 0
				m.end = duration
			}

			// Regenerate view
			if m.waveform != nil {
				view, err := m.waveform.GenerateView(gowaveform.WaveformOptions{
					Start: m.start,
					End:   m.end,
					Width: m.width,
				})
				if err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.currentView = view
			}

		case "right":
			// Jog right - shift the view later in time
			duration := m.end - m.start
			step := duration * 0.005 // Move 0.5% of current view

			m.start += step
			m.end += step

			// Clamp to valid range
			if m.end > m.totalDuration {
				m.end = m.totalDuration
				m.start = m.end - duration
				if m.start < 0 {
					m.start = 0
				}
			}

			// Regenerate view
			if m.waveform != nil {
				view, err := m.waveform.GenerateView(gowaveform.WaveformOptions{
					Start: m.start,
					End:   m.end,
					Width: m.width,
				})
				if err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.currentView = view
			}

		case "up":
			// Zoom in - make start and end closer together
			duration := m.end - m.start
			center := (m.start + m.end) / 2.0
			newDuration := duration * 0.8 // Zoom in by 20%

			m.start = center - newDuration/2.0
			m.end = center + newDuration/2.0

			// Clamp to valid range
			if m.start < 0 {
				m.start = 0
				m.end = newDuration
			}
			if m.end > m.totalDuration {
				m.end = m.totalDuration
				m.start = m.end - newDuration
				if m.start < 0 {
					m.start = 0
				}
			}

			// Regenerate view
			if m.waveform != nil {
				view, err := m.waveform.GenerateView(gowaveform.WaveformOptions{
					Start: m.start,
					End:   m.end,
					Width: m.width,
				})
				if err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.currentView = view
			}

		case "down":
			// Zoom out - make start and end further apart
			duration := m.end - m.start
			center := (m.start + m.end) / 2.0
			newDuration := duration * 1.25 // Zoom out by 25%

			// Don't zoom out beyond total duration
			if newDuration > m.totalDuration {
				newDuration = m.totalDuration
			}

			m.start = center - newDuration/2.0
			m.end = center + newDuration/2.0

			// Clamp to valid range
			if m.start < 0 {
				m.start = 0
				m.end = newDuration
			}
			if m.end > m.totalDuration {
				m.end = m.totalDuration
				m.start = m.end - newDuration
				if m.start < 0 {
					m.start = 0
				}
			}

			// Regenerate view
			if m.waveform != nil {
				view, err := m.waveform.GenerateView(gowaveform.WaveformOptions{
					Start: m.start,
					End:   m.end,
					Width: m.width,
				})
				if err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.currentView = view
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.\n", m.err)
	}

	if m.currentView == nil {
		return "Loading waveform...\n"
	}

	var sb strings.Builder

	// Draw the waveform
	waveformStr := renderWaveform(m.currentView, m.width, m.height-6, m.start, m.end)
	sb.WriteString(waveformStr)
	sb.WriteString("\n")

	// Display information
	sb.WriteString(fmt.Sprintf("File: %s | Duration: %.2fs | Viewing: %.2fs - %.2fs (%.2fs)\n",
		m.wavFile, m.totalDuration, m.start, m.end, m.end-m.start))
	sb.WriteString("Controls: ← → (jog) | ↑ ↓ (zoom) | q (quit)\n")

	return sb.String()
}

// renderWaveform renders the waveform data as ASCII art with timestamp ruler
func renderWaveform(data *gowaveform.WaveformData, width, height int, start, end float64) string {
	if data == nil || len(data.Data) == 0 {
		return "No waveform data"
	}

	// Create a 2D grid for the waveform
	grid := make([][]bool, height)
	for i := range grid {
		grid[i] = make([]bool, width)
	}

	// Find the maximum absolute value for normalization
	var maxAbs int16
	for _, val := range data.Data {
		if val < 0 {
			if -val > maxAbs {
				maxAbs = -val
			}
		} else {
			if val > maxAbs {
				maxAbs = val
			}
		}
	}

	if maxAbs == 0 {
		maxAbs = 1 // Prevent division by zero
	}

	// Plot each min/max pair
	for i := 0; i < len(data.Data)/2 && i < width; i++ {
		minVal := data.Data[i*2]
		maxVal := data.Data[i*2+1]

		// Normalize to height
		// Center is at height/2
		center := height / 2

		minY := center - int(float64(minVal)/float64(maxAbs)*float64(center))
		maxY := center - int(float64(maxVal)/float64(maxAbs)*float64(center))

		// Clamp values
		if minY < 0 {
			minY = 0
		}
		if minY >= height {
			minY = height - 1
		}
		if maxY < 0 {
			maxY = 0
		}
		if maxY >= height {
			maxY = height - 1
		}

		// Ensure minY <= maxY (since we're working in screen coordinates)
		if minY > maxY {
			minY, maxY = maxY, minY
		}

		// Fill the column
		for y := minY; y <= maxY; y++ {
			grid[y][i] = true
		}
	}

	// Convert grid to string
	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] {
				sb.WriteString("█")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// Add timestamp ruler
	sb.WriteString(generateTimestampRuler(width, start, end))

	return sb.String()
}

// generateTimestampRuler creates a timestamp ruler below the waveform
func generateTimestampRuler(width int, start, end float64) string {
	duration := end - start

	// Determine the precision based on the duration
	var precision int
	var interval float64

	if duration < 0.1 {
		precision = 4
		interval = 0.01
	} else if duration < 1.0 {
		precision = 3
		interval = 0.05
	} else if duration < 10.0 {
		precision = 2
		interval = 0.5
	} else if duration < 60.0 {
		precision = 1
		interval = 2.0
	} else {
		precision = 0
		interval = 10.0
	}

	// Calculate number of timestamps to show (aim for ~8-12 timestamps)
	numTimestamps := int(duration / interval)
	if numTimestamps < 5 {
		numTimestamps = 5
		interval = duration / float64(numTimestamps)
	} else if numTimestamps > 15 {
		numTimestamps = 12
		interval = duration / float64(numTimestamps)
	}

	var sb strings.Builder

	// Create tick marks line
	tickLine := make([]rune, width)
	for i := range tickLine {
		tickLine[i] = ' '
	}

	// Create timestamp labels
	timestamps := make(map[int]string)

	for i := 0; i <= numTimestamps; i++ {
		time := start + float64(i)*interval
		if time > end {
			time = end
		}

		// Calculate position
		pos := int(float64(width-1) * (time - start) / duration)
		if pos >= 0 && pos < width {
			tickLine[pos] = '|'

			// Format timestamp based on precision
			var label string
			if precision == 0 {
				label = fmt.Sprintf("%.0f", time)
			} else {
				label = fmt.Sprintf("%.*f", precision, time)
			}
			timestamps[pos] = label
		}
	}

	// Write tick line
	sb.WriteString(string(tickLine))
	sb.WriteString("\n")

	// Write timestamp labels
	labelLine := make([]rune, width)
	for i := range labelLine {
		labelLine[i] = ' '
	}

	for pos, label := range timestamps {
		// Center the label on the tick mark
		startPos := pos - len(label)/2
		if startPos < 0 {
			startPos = 0
		}
		if startPos+len(label) > width {
			startPos = width - len(label)
		}

		// Write label
		for i, ch := range label {
			if startPos+i >= 0 && startPos+i < width {
				labelLine[startPos+i] = ch
			}
		}
	}

	sb.WriteString(string(labelLine))
	sb.WriteString("\n")

	return sb.String()
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: waveform <wav-file>")
		os.Exit(1)
	}

	wavFile := os.Args[1]

	// Check if file exists
	if _, err := os.Stat(wavFile); os.IsNotExist(err) {
		fmt.Printf("Error: File not found: %s\n", wavFile)
		os.Exit(1)
	}

	p := tea.NewProgram(
		initialModel(wavFile),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
