# gowaveform

Pure Go implementation for generating waveform JSON data from WAV files, compatible with [audiowaveform](https://codeberg.org/chrisn/audiowaveform) JSON format.

## Features

- Read WAV files (PCM format, 8-bit and 16-bit) using [go-audio/wav](https://github.com/go-audio/wav)
- Generate waveform data with configurable zoom levels (samples per pixel)
- Support for arbitrary start and end times
- JSON output compatible with audiowaveform format
- Pure Go implementation

## Installation

```bash
go get github.com/schollz/gowaveform
```

To install the CLI tool:

```bash
go install github.com/schollz/gowaveform/cmd/gowaveform@latest
```

## Usage

### As a Library

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

### As a CLI Tool

```bash
# Generate waveform for entire file
gowaveform -i input.wav -o output.json

# Generate waveform for specific time range with custom zoom
gowaveform -i input.wav -o output.json -start 5.0 -end 15.0 -z 512

# Output to stdout
gowaveform -i input.wav

# Custom zoom level (samples per pixel)
gowaveform -i input.wav -o output.json -z 1024
```

### CLI Options

- `-i` : Input WAV file (required)
- `-o` : Output JSON file (default: stdout)
- `-start` : Start time in seconds (default: 0)
- `-end` : End time in seconds (default: end of file)
- `-z` : Zoom level, samples per pixel (default: 256)

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

- **Audio Format**: PCM (uncompressed)
- **Bit Depths**: 8-bit, 16-bit
- **Channels**: Mono and stereo

## Example

```bash
# Create a waveform from a 60-second audio file, zoomed to show detail
gowaveform -i song.wav -o waveform.json -z 512

# Generate waveform for a 10-second segment
gowaveform -i podcast.wav -o segment.json -start 30 -end 40 -z 256
```

## License

MIT