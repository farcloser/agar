/*
   Copyright Mycophonic.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package agar

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"time"
)

// Default benchmark parameters.
const (
	DefaultBenchIterations = 10
	DefaultBenchDuration   = 10 * time.Second
)

// BenchFormat describes an audio format configuration for benchmarking.
type BenchFormat struct {
	Name       string
	SampleRate int
	BitDepth   int
	Channels   int
}

// BenchOptions controls benchmark execution parameters.
// Zero values are replaced with defaults by WithDefaults.
type BenchOptions struct {
	Iterations int
	Duration   time.Duration
}

// WithDefaults returns a copy with zero fields replaced by defaults.
func (o BenchOptions) WithDefaults() BenchOptions {
	if o.Iterations == 0 {
		o.Iterations = DefaultBenchIterations
	}

	if o.Duration == 0 {
		o.Duration = DefaultBenchDuration
	}

	return o
}

// DurationSeconds returns the audio duration as whole seconds.
func (o BenchOptions) DurationSeconds() int {
	return int(o.Duration.Seconds())
}

// BenchResult holds timing statistics for a single benchmark run.
type BenchResult struct {
	Format  BenchFormat
	Tool    string
	Op      string
	Median  time.Duration
	Mean    time.Duration
	Min     time.Duration
	Max     time.Duration
	Stddev  time.Duration
	PCMSize int
}

// ComputeResult calculates timing statistics from a set of measured durations.
func ComputeResult(format BenchFormat, tool, operation string, durations []time.Duration, pcmSize int) BenchResult {
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	slices.Sort(sorted)

	var sum float64
	for _, d := range durations {
		sum += float64(d)
	}

	mean := sum / float64(len(durations))

	var variance float64

	for _, d := range durations {
		diff := float64(d) - mean
		variance += diff * diff
	}

	variance /= float64(len(durations))

	return BenchResult{
		Format:  format,
		Tool:    tool,
		Op:      operation,
		Median:  sorted[len(sorted)/2],
		Mean:    time.Duration(mean),
		Min:     sorted[0],
		Max:     sorted[len(sorted)-1],
		Stddev:  time.Duration(math.Sqrt(variance)),
		PCMSize: pcmSize,
	}
}

// PrintResults displays benchmark results in a formatted table.
func PrintResults(t *testing.T, opts BenchOptions, results []BenchResult) {
	t.Helper()

	opts = opts.WithDefaults()

	sep := "─────────────────────────────────────────────────────────────────────────────"

	t.Log("")
	t.Log("┌" + sep + "┐")
	t.Logf("│ Benchmark Results%59s│", "")
	t.Logf("│ Iterations: %-4d   Audio duration: %-4ds%35s│", opts.Iterations, opts.DurationSeconds(), "")
	t.Log("├" + sep + "┤")
	t.Logf("│ %-28s %-12s %-6s %8s %8s %8s %8s│", "Format", "Tool", "Op", "Median", "Mean", "StdDev", "Min")
	t.Log("├" + sep + "┤")

	currentName := ""

	for _, result := range results {
		label := formatLabel(result.Format)

		if label != currentName {
			if currentName != "" {
				t.Log("├" + sep + "┤")
			}

			currentName = label

			t.Logf("│ %-76s│", fmt.Sprintf("%s  (%dHz %dbit %dch, PCM %.1f MB)",
				result.Format.Name, result.Format.SampleRate, result.Format.BitDepth, result.Format.Channels,
				float64(result.PCMSize)/(1024*1024)))
		}

		t.Logf("│ %-28s %-12s %-6s %8s %8s %8s %8s│",
			"", result.Tool, result.Op,
			result.Median.Round(time.Millisecond),
			result.Mean.Round(time.Millisecond),
			result.Stddev.Round(time.Millisecond),
			result.Min.Round(time.Millisecond),
		)
	}

	t.Log("└" + sep + "┘")
}

func formatLabel(format BenchFormat) string {
	return fmt.Sprintf("%s_%d_%d_%d", format.Name, format.SampleRate, format.BitDepth, format.Channels)
}
