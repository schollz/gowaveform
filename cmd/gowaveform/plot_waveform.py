#!/usr/bin/env python3
"""
Plot waveform data using Plotly.
Reads JSON data from stdin or a file and creates an interactive waveform visualization.
"""

import sys
import json
import plotly.graph_objects as go


def plot_waveform(data):
    """
    Plot waveform data using Plotly.
    
    Args:
        data: Dictionary containing waveform data in audiowaveform format
    """
    # Extract metadata
    sample_rate = data.get('sample_rate', 44100)
    samples_per_pixel = data.get('samples_per_pixel', 256)
    channels = data.get('channels', 1)
    waveform_data = data.get('data', [])
    
    # Calculate time axis
    # Each pixel represents samples_per_pixel samples at the given sample rate
    time_per_pixel = samples_per_pixel / sample_rate
    num_pixels = len(waveform_data) // 2
    time_axis = [i * time_per_pixel for i in range(num_pixels)]
    
    # Extract min and max values
    # Data format: [min1, max1, min2, max2, ...]
    min_values = [waveform_data[i] for i in range(0, len(waveform_data), 2)]
    max_values = [waveform_data[i] for i in range(1, len(waveform_data), 2)]
    
    # Create the waveform plot
    fig = go.Figure()
    
    # Add max values (upper envelope)
    fig.add_trace(go.Scatter(
        x=time_axis,
        y=max_values,
        mode='lines',
        name='Max',
        line=dict(color='rgba(0, 100, 200, 0.5)', width=0.5),
        fill=None,
        showlegend=False
    ))
    
    # Add min values (lower envelope) with fill to create the waveform shape
    fig.add_trace(go.Scatter(
        x=time_axis,
        y=min_values,
        mode='lines',
        name='Min',
        line=dict(color='rgba(0, 100, 200, 0.5)', width=0.5),
        fill='tonexty',
        fillcolor='rgba(0, 100, 200, 0.3)',
        showlegend=False
    ))
    
    # Update layout
    fig.update_layout(
        title=f'Waveform Visualization<br><sub>Sample Rate: {sample_rate} Hz | '
              f'Samples per Pixel: {samples_per_pixel} | Channels: {channels}</sub>',
        xaxis_title='Time (seconds)',
        yaxis_title='Amplitude',
        hovermode='x unified',
        template='plotly_white',
        height=500,
        xaxis=dict(
            showgrid=True,
            gridcolor='lightgray',
        ),
        yaxis=dict(
            showgrid=True,
            gridcolor='lightgray',
            zeroline=True,
            zerolinecolor='black',
            zerolinewidth=1,
        )
    )
    
    # Show the plot
    fig.show()


def main():
    """Main function to read JSON and plot waveform."""
    if len(sys.argv) > 1:
        # Read from file
        filename = sys.argv[1]
        try:
            with open(filename, 'r') as f:
                data = json.load(f)
        except FileNotFoundError:
            print(f"Error: File '{filename}' not found", file=sys.stderr)
            sys.exit(1)
        except json.JSONDecodeError as e:
            print(f"Error: Invalid JSON in file '{filename}': {e}", file=sys.stderr)
            sys.exit(1)
    else:
        # Read from stdin
        try:
            data = json.load(sys.stdin)
        except json.JSONDecodeError as e:
            print(f"Error: Invalid JSON from stdin: {e}", file=sys.stderr)
            sys.exit(1)
    
    # Validate data structure
    if not isinstance(data, dict):
        print("Error: JSON data must be an object", file=sys.stderr)
        sys.exit(1)
    
    if 'data' not in data:
        print("Error: JSON data must contain 'data' field", file=sys.stderr)
        sys.exit(1)
    
    if not isinstance(data['data'], list):
        print("Error: 'data' field must be an array", file=sys.stderr)
        sys.exit(1)
    
    if len(data['data']) == 0:
        print("Error: 'data' field is empty", file=sys.stderr)
        sys.exit(1)
    
    # Plot the waveform
    plot_waveform(data)


if __name__ == '__main__':
    main()
