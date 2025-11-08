package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/schollz/gowaveform"
)

func main() {
	// Command line flags
	inputFile := flag.String("i", "", "Input WAV file (required)")
	outputFile := flag.String("o", "", "Output JSON file (default: stdout)")
	start := flag.Float64("start", 0, "Start time in seconds (default: 0)")
	end := flag.Float64("end", 0, "End time in seconds (default: end of file)")
	zoom := flag.Int("z", 256, "Zoom level (samples per pixel, default: 256)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gowaveform [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generate waveform JSON data from WAV files.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  gowaveform -i input.wav -o output.json -start 0 -end 10 -z 512\n")
	}

	flag.Parse()

	// Validate input
	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: input file '%s' does not exist\n", *inputFile)
		os.Exit(1)
	}

	// Generate waveform data
	opts := gowaveform.WaveformOptions{
		Start:           *start,
		End:             *end,
		SamplesPerPixel: *zoom,
	}

	jsonData, err := gowaveform.GenerateWaveformJSON(*inputFile, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating waveform: %v\n", err)
		os.Exit(1)
	}

	// Output JSON
	if *outputFile == "" {
		// Write to stdout
		fmt.Println(string(jsonData))
	} else {
		// Write to file
		if err := os.WriteFile(*outputFile, jsonData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Waveform data written to %s\n", *outputFile)
	}
}
