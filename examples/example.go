package main

import (
	"fmt"
	"log"
	"os"

	"github.com/schollz/gowaveform"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run example.go <input.wav>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// Example 1: Generate waveform for entire file
	fmt.Println("Generating waveform for entire file...")
	opts := gowaveform.WaveformOptions{
		Start:           0,
		End:             0, // 0 means end of file
		SamplesPerPixel: 256,
	}

	data, err := gowaveform.GenerateWaveformData(inputFile, opts)
	if err != nil {
		log.Fatalf("Error generating waveform: %v", err)
	}

	fmt.Printf("Generated waveform with %d pixels\n", data.Length)
	fmt.Printf("Sample rate: %d Hz\n", data.SampleRate)
	fmt.Printf("Channels: %d\n", data.Channels)
	fmt.Printf("Samples per pixel: %d\n", data.SamplesPerPixel)

	// Example 2: Generate JSON
	fmt.Println("\nGenerating JSON output...")
	jsonData, err := gowaveform.GenerateJSON(data)
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	// Write to file
	outputFile := "waveform.json"
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
	fmt.Printf("Waveform data written to %s\n", outputFile)

	// Example 3: Generate waveform for a specific time range
	fmt.Println("\nGenerating waveform for time range (0.5-1.5 seconds)...")
	rangeOpts := gowaveform.WaveformOptions{
		Start:           0.5,
		End:             1.5,
		SamplesPerPixel: 512,
	}

	rangeData, err := gowaveform.GenerateWaveformData(inputFile, rangeOpts)
	if err != nil {
		log.Fatalf("Error generating range waveform: %v", err)
	}

	fmt.Printf("Generated waveform with %d pixels for 1 second range\n", rangeData.Length)

	rangeJSON, err := gowaveform.GenerateJSON(rangeData)
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	rangeOutputFile := "waveform_range.json"
	if err := os.WriteFile(rangeOutputFile, rangeJSON, 0644); err != nil {
		log.Fatalf("Error writing file: %v", err)
	}
	fmt.Printf("Range waveform data written to %s\n", rangeOutputFile)

	// Show first few data points
	fmt.Println("\nFirst 10 min/max pairs from the waveform:")
	for i := 0; i < 10 && i*2+1 < len(data.Data); i++ {
		fmt.Printf("Pixel %d: min=%d, max=%d\n", i, data.Data[i*2], data.Data[i*2+1])
	}
}
