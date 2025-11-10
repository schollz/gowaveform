# gowaveform

[![CI](https://github.com/schollz/gowaveform/actions/workflows/CI.yml/badge.svg)](https://github.com/schollz/gowaveform/actions/workflows/CI.yml)
[![codecov](https://codecov.io/gh/schollz/gowaveform/branch/main/graph/badge.svg)](https://codecov.io/gh/schollz/gowaveform)
[![Go Reference](https://pkg.go.dev/badge/github.com/schollz/gowaveform.svg)](https://pkg.go.dev/github.com/schollz/gowaveform)
[![Release](https://img.shields.io/github/v/release/schollz/gowaveform.svg)](https://github.com/schollz/gowaveform/releases/latest)

Pure Go implementation for generating waveform JSON data from WAV files, compatible with [audiowaveform](https://codeberg.org/chrisn/audiowaveform) JSON format.

<img width="800" height="400" alt="1" src="https://github.com/user-attachments/assets/ad488d2e-e576-421a-8c87-b5f9422e4a14" />


## Features

- Read audio files (WAV, MP3, FLAC, OGG, etc.) using [audiomorph](https://github.com/schollz/audiomorph)
- Generate waveform data with configurable zoom levels (samples per pixel)
- Support for arbitrary start and end times
- JSON output compatible with audiowaveform format
- **Save waveform visualizations as PNG or JPEG images** with customizable dimensions and colors
- Pure Go implementation

## Installation

```bash
go get github.com/schollz/gowaveform
```

To install the terminal-based waveform visualizer:

```bash
go install github.com/schollz/gowaveform/cmd/gowaveform@latest
```

## Usage

### As a Library

#### Generate JSON Waveform Data

```go
package main

import (
    "fmt"
    "github.com/schollz/gowaveform"
)

func main() {
    opts := gowaveform.WaveformOptions{
        Start:           0.0,   // Start time in seconds
        End:             10.0,  // End time in seconds (0 = end of file)
        SamplesPerPixel: 256,   // Zoom level
    }
    
    jsonData, err := gowaveform.GenerateWaveformJSON("input.wav", opts)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(jsonData))
}
```

#### Save Waveform as Image

You can save waveform visualizations as PNG or JPEG images using the plot API:

```go
package main

import (
    "log"
    "github.com/schollz/gowaveform"
)

func main() {
    // Load the waveform
    waveform, err := gowaveform.LoadWaveform("audio.wav")
    if err != nil {
        log.Fatal(err)
    }

    // Save as PNG with custom options
    err = gowaveform.SavePlot(waveform, "output.png",
        gowaveform.OptionSetWidth(1200),
        gowaveform.OptionSetHeight(400),
        gowaveform.OptionSetBackgroundColor("#FFFFFF"),
        gowaveform.OptionSetForegroundColor("#0064C8"),
        gowaveform.OptionShowTimestamp(true),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

**Available Options:**
- `OptionSetWidth(width int)` - Set plot width in pixels (default: 800)
- `OptionSetHeight(height int)` - Set plot height in pixels (default: 400)
- `OptionSetBackgroundColor(hexColor string)` - Set background color (e.g., "#FFFFFF")
- `OptionSetForegroundColor(hexColor string)` - Set waveform color (e.g., "#0064C8")
- `OptionShowTimestamp(show bool)` - Enable/disable time axis (default: true)

The file format (PNG or JPEG) is determined by the filename extension.

### Command-Line Tool

The CLI tool can be used in two modes: interactive visualization or direct image generation.

#### Generate Waveform Images

Generate waveform plots directly to image files without launching the interactive viewer:

```bash
# Generate a PNG with default settings (800x400)
gowaveform audio.wav --output waveform.png

# Generate a JPEG with custom dimensions
gowaveform audio.wav --output waveform.jpg --width 1200 --height 400

# Customize colors (hex format)
gowaveform audio.wav --output waveform.png \
  --bg-color "#FFFFFF" --fg-color "#0064C8"

# Generate without timestamp axis
gowaveform audio.wav --output waveform.png --no-timestamp

# Combine all options
gowaveform audio.wav --output waveform.png \
  --width 1600 --height 800 \
  --bg-color "#1a1a1a" --fg-color "#00d4ff" \
  --no-timestamp
```

**Available Flags:**
- `--output`, `-o` - Output file path (PNG or JPEG format)
- `--width` - Width of the plot in pixels (default: 800)
- `--height` - Height of the plot in pixels (default: 400)
- `--bg-color` - Background color in hex format (e.g., "#FFFFFF")
- `--fg-color` - Foreground/waveform color in hex format (e.g., "#0064C8")
- `--no-timestamp` - Disable timestamp axis on the plot

#### Interactive Visualizer

Launch the interactive terminal-based waveform visualizer for navigating, zooming, and marking positions in WAV files:

```bash
# Launch the interactive visualizer
gowaveform audio.wav
```

**Controls:**
- `m` / `Space` - Create marker at center of view
- `o` - Run onset detection and create markers
- `Tab` - Cycle through slices
- `Shift+Tab` - Cycle through markers
- `d` / `Backspace` - Delete selected marker/slice
- `e` - Export slices to JSON
- `Esc` - Unselect marker/slice
- `←` / `→` - Jog view or selected marker
- `Shift+←` / `Shift+→` - Fast jog view
- `↑` / `↓` - Zoom in/out
- `q` - Quit

## JSON Output Format

The output JSON follows the audiowaveform format:

```json
{
  "version": 2,
  "channels": 1,
  "sample_rate": 44100,
  "samples_per_pixel": 256,
  "bits": 16,
  "length": 173,
  "data": [
    -100, 100,
    -120, 95,
    ...
  ]
}
```

The `data` array contains min/max pairs for each pixel, allowing visualization programs to render the waveform.

## Supported Formats

audiomorph supports a wide variety of audio formats including:
- **Audio Formats**: WAV, MP3, FLAC, OGG, AIFF
- **Bit Depths**: 8-bit, 16-bit, 24-bit, 32-bit
- **Channels**: Mono and multi-channel audio

## Example

```bash
# Generate waveform JSON data programmatically
gowaveform.GenerateWaveformJSON("song.wav", gowaveform.WaveformOptions{
    Start: 0,
    End: 60,
    SamplesPerPixel: 512,
})

# Generate a waveform image from the command line
gowaveform song.wav --output waveform.png

# Use the interactive terminal visualizer
gowaveform song.wav
```

## License

MIT
