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

		case "left", "shift+left":
			// Jog left - shift the view earlier in time
			duration := m.end - m.start
			step := duration * 0.005 // Move 0.5% of current view (default)

			if msg.String() == "shift+left" {
				step = duration * 0.05 // Move 5% of current view (faster)
			}

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

		case "right", "shift+right":
			// Jog right - shift the view later in time
			duration := m.end - m.start
			step := duration * 0.005 // Move 0.5% of current view (default)

			if msg.String() == "shift+right" {
				step = duration * 0.05 // Move 5% of current view (faster)
			}

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
	sb.WriteString("Controls: ← → (jog) | Shift+← → (fast jog) | ↑ ↓ (zoom) | q (quit)\n")

	return sb.String()
}

// renderWaveform renders the waveform data as high-resolution art using Unicode block characters
func renderWaveform(data *gowaveform.WaveformData, width, height int, start, end float64) string {
	if data == nil || len(data.Data) == 0 {
		return "No waveform data"
	}

	// Use 8 vertical segments per character for higher resolution
	// This means we multiply height by 8 for our internal grid
	const segmentsPerChar = 8
	virtualHeight := height * segmentsPerChar

	// Create a higher resolution grid (8 segments per character height)
	grid := make([][]bool, virtualHeight)
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

		// Normalize to virtual height
		center := virtualHeight / 2

		minY := center - int(float64(minVal)/float64(maxAbs)*float64(center))
		maxY := center - int(float64(maxVal)/float64(maxAbs)*float64(center))

		// Clamp values
		if minY < 0 {
			minY = 0
		}
		if minY >= virtualHeight {
			minY = virtualHeight - 1
		}
		if maxY < 0 {
			maxY = 0
		}
		if maxY >= virtualHeight {
			maxY = virtualHeight - 1
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

	// Convert high-resolution grid to block characters
	// Split rendering into upper and lower halves for proper block usage
	var sb strings.Builder
	centerY := height / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Determine if we're in upper or lower half
			if y < centerY {
				// Upper half: use lower blocks inverted (hanging from top of cell)
				char := getUpperHalfChar(grid, x, y, segmentsPerChar)
				sb.WriteString(char)
			} else {
				// Lower half: use upper blocks (extending from bottom of cell)
				char := getLowerHalfChar(grid, x, y, segmentsPerChar)
				sb.WriteString(char)
			}
		}
		sb.WriteString("\n")
	}

	// Add timestamp ruler
	sb.WriteString(generateTimestampRuler(width, start, end))

	return sb.String()
}

// getUpperHalfChar returns block character for upper half of waveform
// Uses upper blocks (measuring down from top of character cell)
func getUpperHalfChar(grid [][]bool, x, y, segmentsPerChar int) string {
	baseY := y * segmentsPerChar

	// Find the lowest filled segment (deepest extent into this cell from top)
	lowestFilled := -1
	for i := segmentsPerChar - 1; i >= 0; i-- {
		segY := baseY + i
		if segY < len(grid) && grid[segY][x] {
			lowestFilled = i
			break
		}
	}

	// If nothing filled, return empty
	if lowestFilled == -1 {
		return " "
	}

	// Use upper blocks that hang from the top
	// lowestFilled ranges from 0 (top) to 7 (bottom of cell)
	// Upper blocks fill from top, so we map based on extent
	extent := lowestFilled + 1 // +1 because index 0 means 1 segment filled

	switch extent {
	case 1:
		return "▔" // U+2594 Upper one eighth
	case 2:
		return "🮂" // U+1FB02 Upper one quarter
	case 3:
		return "🮃" // U+1FB03 Upper three eighths
	case 4:
		return "▀" // U+2580 Upper half
	case 5:
		return "🮄" // U+1FB04 Upper five eighths
	case 6:
		return "🮅" // U+1FB05 Upper three quarters
	case 7:
		return "🮆" // U+1FB06 Upper seven eighths
	default: // 8
		return "█" // U+2588 - full cell
	}
}

// getLowerHalfChar returns block character for lower half of waveform
// Uses lower blocks (measuring up from bottom of character cell)
func getLowerHalfChar(grid [][]bool, x, y, segmentsPerChar int) string {
	baseY := y * segmentsPerChar

	// Find the highest filled segment (highest extent into this cell from bottom)
	highestFilled := -1
	for i := 0; i < segmentsPerChar; i++ {
		segY := baseY + i
		if segY < len(grid) && grid[segY][x] {
			highestFilled = i
			break
		}
	}

	// If nothing filled, return empty
	if highestFilled == -1 {
		return " "
	}

	// Use lower blocks that extend from the bottom
	// highestFilled ranges from 0 (top of cell) to 7 (bottom of cell)
	// Lower blocks fill from bottom, so we need to invert
	// If segment 0 (top) is filled, we need a full or near-full block
	// If segment 7 (bottom) is filled, we need just a small bottom block
	extent := segmentsPerChar - highestFilled

	switch extent {
	case 1:
		return "▁" // U+2581 - one eighth from bottom
	case 2:
		return "▂" // U+2582
	case 3:
		return "▃" // U+2583
	case 4:
		return "▄" // U+2584
	case 5:
		return "▅" // U+2585
	case 6:
		return "▆" // U+2586
	case 7:
		return "▇" // U+2587
	default: // 8
		return "█" // U+2588 - full cell
	}
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
