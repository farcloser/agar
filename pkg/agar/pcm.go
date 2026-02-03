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
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// Bit depth constants for PCM audio formats.
// BitDepth8, BitDepth24, BitDepth32 are defined in ffmpeg.go.
const (
	BitDepth4  = 4
	BitDepth12 = 12
	BitDepth16 = 16
	BitDepth20 = 20
)

// xorshift64 PRNG parameters (Marsaglia's constants).
const (
	xorshiftSeed   = uint64(0x12345678)
	xorshiftShiftA = 13
	xorshiftShiftB = 7
	xorshiftShiftC = 17
)

// Comparison thresholds.
const (
	defaultMaxDiffSamples = 5
	lossyLargeDiffPct     = 100
)

// PCMBytesPerSample returns the number of bytes per sample for a given bit depth.
func PCMBytesPerSample(bitDepth int) int {
	switch bitDepth {
	case BitDepth4, BitDepth8:
		return 1
	case BitDepth12, BitDepth16:
		return 2
	case BitDepth20, BitDepth24:
		return 3
	case BitDepth32:
		return 4
	default:
		return bitDepth / bitsPerByte
	}
}

// GenerateWhiteNoise creates deterministic random PCM data at the given format.
// The PRNG is seeded with a fixed value so output is reproducible across runs.
func GenerateWhiteNoise(sampleRate, bitDepth, channels, durationSec int) []byte {
	numSamples := sampleRate * durationSec * channels
	bytesPerSample := PCMBytesPerSample(bitDepth)

	buf := make([]byte, numSamples*bytesPerSample)

	seed := xorshiftSeed

	for sampleIdx := range numSamples {
		seed ^= seed << xorshiftShiftA
		seed ^= seed >> xorshiftShiftB
		seed ^= seed << xorshiftShiftC

		offset := sampleIdx * bytesPerSample

		switch bitDepth {
		case BitDepth4:
			buf[offset] = byte(int8((seed % 14) - 7)) //nolint:gosec // G115: modulo bounds result to -7..+6, fits int8.
		case BitDepth8:
			buf[offset] = byte(
				int8((seed % 240) - 120), //nolint:gosec // G115: modulo bounds result to -120..+119, fits int8.
			)
		case BitDepth12:
			val := int16( //nolint:gosec // G115: modulo bounds result to -2000..+1999, fits int16.
				(seed % 4000) - 2000,
			)
			binary.LittleEndian.PutUint16(
				buf[offset:],
				uint16(val), //nolint:gosec // G115: reinterpret cast for LE encoding.
			)
		case BitDepth16:
			val := int16( //nolint:gosec // G115: modulo bounds result to -30000..+29999, fits int16.
				(seed % 60000) - 30000,
			)
			binary.LittleEndian.PutUint16(
				buf[offset:],
				uint16(val), //nolint:gosec // G115: reinterpret cast for LE encoding.
			)
		case BitDepth20:
			val := int32( //nolint:gosec // G115: modulo bounds result to ±500000, fits int32.
				(seed % 1000000) - 500000,
			)
			buf[offset] = byte(val)
			buf[offset+1] = byte(val >> bitsPerByte)
			buf[offset+2] = byte(val >> (2 * bitsPerByte))
		case BitDepth24:
			val := int32( //nolint:gosec // G115: modulo bounds result to ±7000000, fits int32.
				(seed % 14000000) - 7000000,
			)
			buf[offset] = byte(val)
			buf[offset+1] = byte(val >> bitsPerByte)
			buf[offset+2] = byte(val >> (2 * bitsPerByte))
		case BitDepth32:
			val := int32( //nolint:gosec // G115: modulo bounds result to ±900M, fits int32.
				(seed % 1800000000) - 900000000,
			)
			binary.LittleEndian.PutUint32(
				buf[offset:],
				uint32(val), //nolint:gosec // G115: reinterpret cast for LE encoding.
			)

		default:
		}
	}

	return buf
}

// CompareLosslessSamples requires exact byte match for lossless codecs.
// The label identifies which comparison is being made (e.g. "saprobe vs ffmpeg").
func CompareLosslessSamples(t *testing.T, label string, expected, actual []byte, bitDepth, channels int) {
	t.Helper()

	minLen := min(len(expected), len(actual))
	differences := 0
	firstDiff := -1

	for idx := range minLen {
		if expected[idx] != actual[idx] {
			differences++

			if firstDiff == -1 {
				firstDiff = idx
			}
		}
	}

	if differences > 0 {
		bytesPerSample := PCMBytesPerSample(bitDepth)
		sampleIndex := firstDiff / bytesPerSample / channels
		t.Errorf("%s: PCM mismatch: %d differing bytes (%.2f%%), first diff at byte %d (sample %d)",
			label, differences, float64(differences)/float64(minLen)*lossyLargeDiffPct, firstDiff, sampleIndex)

		ShowDiffs(t, label, expected, actual, bitDepth, channels, defaultMaxDiffSamples)
	}
}

// CompareLossySamples allows small differences between decoders for lossy codecs.
// Different decoders use different floating-point implementations,
// resulting in +/-1-2 LSB differences per sample.
// Length differences up to 1 frame (1152 stereo samples) are tolerated.
func CompareLossySamples(t *testing.T, pcmA, pcmB []byte, bitDepth, channels int) {
	t.Helper()

	if bitDepth != BitDepth16 {
		t.Errorf("lossy comparison only supports 16-bit, got %d-bit", bitDepth)

		return
	}

	// Allow length differences up to 1 frame (1152 samples * channels * 2 bytes).
	const samplesPerFrame = 1152

	maxLengthDiffBytes := samplesPerFrame * channels * 2

	lengthDiff := len(pcmA) - len(pcmB)
	if lengthDiff < 0 {
		lengthDiff = -lengthDiff
	}

	if lengthDiff > maxLengthDiffBytes {
		t.Errorf("length mismatch: a=%d, b=%d (diff=%d exceeds tolerance %d)",
			len(pcmA), len(pcmB), lengthDiff, maxLengthDiffBytes)

		return
	}

	if lengthDiff > 0 {
		t.Logf("length diff: a=%d, b=%d (+/-%d bytes, within tolerance)",
			len(pcmA), len(pcmB), lengthDiff)
	}

	numSamples := min(len(pcmA), len(pcmB)) / 2

	const maxDiffPerSample = 2

	largeDiffs := 0
	maxDiff := int16(0)

	for idx := range numSamples {
		sampleA := int16( //nolint:gosec // G115: reinterpret uint16 as signed for PCM comparison.
			binary.LittleEndian.Uint16(pcmA[idx*2:]),
		)
		sampleB := int16( //nolint:gosec // G115: reinterpret uint16 as signed for PCM comparison.
			binary.LittleEndian.Uint16(pcmB[idx*2:]),
		)

		diff := sampleA - sampleB
		if diff < 0 {
			diff = -diff
		}

		if diff > maxDiffPerSample {
			largeDiffs++
		}

		if diff > maxDiff {
			maxDiff = diff
		}
	}

	// Allow up to 1% of samples to have larger differences (codec edge cases).
	maxLargeDiffs := numSamples / lossyLargeDiffPct
	if largeDiffs > maxLargeDiffs {
		t.Errorf("lossy PCM mismatch: %d samples (%.2f%%) differ by more than +/-%d, max diff=%d",
			largeDiffs, float64(largeDiffs)/float64(numSamples)*lossyLargeDiffPct, maxDiffPerSample, maxDiff)
		ShowDiffs(t, "lossy comparison", pcmA, pcmB, bitDepth, channels, defaultMaxDiffSamples)
	}
}

// ShowDiffs prints the first maxDiffs differing frames for debugging.
func ShowDiffs(t *testing.T, label string, expected, actual []byte, bitDepth, channels, maxDiffs int) {
	t.Helper()

	bytesPerSample := PCMBytesPerSample(bitDepth)
	frameSize := bytesPerSample * channels
	shown := 0

	for idx := 0; idx < min(len(expected), len(actual))-frameSize && shown < maxDiffs; idx += frameSize {
		expectedFrame := expected[idx : idx+frameSize]
		actualFrame := actual[idx : idx+frameSize]

		if !bytes.Equal(expectedFrame, actualFrame) {
			sampleIdx := idx / frameSize
			t.Logf("%s: sample %d: expected=%v, actual=%v", label, sampleIdx, expectedFrame, actualFrame)

			shown++
		}
	}
}

// MonoToStereo duplicates mono PCM to stereo by copying each sample to both channels.
func MonoToStereo(mono []byte, bitDepth int) []byte {
	bps := PCMBytesPerSample(bitDepth)
	numSamples := len(mono) / bps
	stereo := make([]byte, numSamples*bps*2)

	for idx := range numSamples {
		sample := mono[idx*bps : (idx+1)*bps]
		copy(stereo[idx*bps*2:], sample)
		copy(stereo[idx*bps*2+bps:], sample)
	}

	return stereo
}

// FileSize returns the size of the file at path, or fatals the test.
func FileSize(t *testing.T, path string) int {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	return int(info.Size())
}

// ProjectRoot walks up from the current working directory to find go.mod.
func ProjectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}

		dir = parent
	}
}
