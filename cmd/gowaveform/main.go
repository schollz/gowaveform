package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/schollz/gowaveform"
)

func main() {
	// Command line flags
	inputFile := flag.String("i", "", "Input WAV file (required)")
	outputFile := flag.String("o", "", "Output JSON file (default: stdout)")
	start := flag.Float64("start", 0, "Start time in seconds (default: 0)")
	end := flag.Float64("end", 0, "End time in seconds (default: end of file)")
	zoom := flag.Int("z", 256, "Zoom level (samples per pixel, default: 256)")
	plot := flag.Bool("plot", false, "Plot waveform using Plotly (requires plotly to be installed)")

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
	if *plot {
		// When plotting, we need to provide the JSON data to the Python script
		// Find the plot_waveform.py script relative to the executable
		exePath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
			os.Exit(1)
		}
		
		// The script is in the cmd directory relative to the executable or source
		scriptPath := filepath.Join(filepath.Dir(exePath), "plot_waveform.py")
		
		// If not found, try relative to source (for development)
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			scriptPath = filepath.Join(filepath.Dir(filepath.Dir(exePath)), "cmd", "plot_waveform.py")
		}
		
		// Still not found? Try from current directory
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			scriptPath = "cmd/plot_waveform.py"
		}
		
		// Final fallback: look in the same directory as the executable
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: plot_waveform.py script not found\n")
			fmt.Fprintf(os.Stderr, "Please ensure plot_waveform.py is in the cmd directory\n")
			os.Exit(1)
		}
		
		// Run the Python script with the JSON data
		cmd := exec.Command("python3", scriptPath)
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		// Create a pipe to write JSON to the script's stdin
		stdin, err := cmd.StdinPipe()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating pipe: %v\n", err)
			os.Exit(1)
		}
		
		// Start the command
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting plot script: %v\n", err)
			fmt.Fprintf(os.Stderr, "Make sure Python 3 and plotly are installed\n")
			os.Exit(1)
		}
		
		// Write JSON data to the script's stdin
		if _, err := stdin.Write(jsonData); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to plot script: %v\n", err)
			os.Exit(1)
		}
		stdin.Close()
		
		// Wait for the command to finish
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running plot script: %v\n", err)
			os.Exit(1)
		}
		
		// Also write to output file if specified
		if *outputFile != "" {
			if err := os.WriteFile(*outputFile, jsonData, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Waveform data written to %s\n", *outputFile)
		}
	} else if *outputFile == "" {
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
