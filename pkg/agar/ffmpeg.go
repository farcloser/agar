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
	"context"
	"io"
	"os/exec"
	"strconv"
	"testing"
)

// Bit depth constants for PCM format and codec mapping.
const (
	BitDepth8  = 8
	BitDepth24 = 24
	BitDepth32 = 32
)

// FFmpegOptions configures an ffmpeg invocation.
type FFmpegOptions struct {
	// Args are passed directly to the ffmpeg binary.
	Args []string
	// Stdin is connected to the command's standard input when non-nil.
	Stdin io.Reader
	// Stdout receives the command's standard output when non-nil.
	// When nil, stdout is captured and returned in FFmpegResult.Stdout.
	Stdout io.Writer
	// Stderr receives the command's standard error when non-nil.
	// When nil, stderr is captured and included in the fatal message on failure.
	Stderr io.Writer
}

// FFmpegResult holds captured output from an ffmpeg invocation.
type FFmpegResult struct {
	// Stdout contains captured standard output, populated only when
	// FFmpegOptions.Stdout was nil.
	Stdout []byte
}

// FFmpeg runs ffmpeg with the given options.
// It fatals the test if ffmpeg cannot be found or the command returns an error.
func FFmpeg(t *testing.T, opts FFmpegOptions) FFmpegResult {
	t.Helper()

	ffmpegPath, err := LookFor(ffmpegBinary)
	if err != nil {
		t.Log(ffmpegBinary + ": " + err.Error())
		t.FailNow()
	}

	//nolint:gosec // arguments are test-controlled
	cmd := exec.CommandContext(context.Background(), ffmpegPath, opts.Args...)

	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}

	var stdoutBuf bytes.Buffer

	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = &stdoutBuf
	}

	var stderrBuf bytes.Buffer

	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = &stderrBuf
	}

	if err := cmd.Run(); err != nil {
		t.Fatalf("ffmpeg: %v\n%s", err, stderrBuf.String())
	}

	return FFmpegResult{
		Stdout: stdoutBuf.Bytes(),
	}
}

// FFmpegEncodeOptions configures encoding raw PCM to a compressed format.
type FFmpegEncodeOptions struct {
	// Src is the path to the raw PCM input file.
	Src string
	// Dst is the path for the encoded output file.
	Dst string
	// BitDepth of the input PCM (determines the raw format via RawPCMFormat).
	BitDepth int
	// SampleRate of the input PCM (-ar).
	SampleRate int
	// Channels in the input PCM (-ac).
	Channels int
	// CodecArgs are codec selection and options placed after -i.
	// For example: "-c:a", "alac", "-sample_fmt", "s32p".
	CodecArgs []string
	// InputArgs are optional extra arguments placed before -i.
	// For example: "-channel_layout", "5.1".
	InputArgs []string
}

// FFmpegEncode encodes a raw PCM file to the target format.
// It fatals the test if ffmpeg cannot be found or the command returns an error.
func FFmpegEncode(t *testing.T, opts FFmpegEncodeOptions) {
	t.Helper()

	args := []string{
		"-y",
		"-f", RawPCMFormat(opts.BitDepth),
		"-ar", strconv.Itoa(opts.SampleRate),
		"-ac", strconv.Itoa(opts.Channels),
	}

	args = append(args, opts.InputArgs...)
	args = append(args, "-i", opts.Src)
	args = append(args, opts.CodecArgs...)
	args = append(args, opts.Dst)

	FFmpeg(t, FFmpegOptions{Args: args})
}

// FFmpegDecodeOptions configures decoding an audio file to raw PCM.
type FFmpegDecodeOptions struct {
	// Src is the path to the encoded input file.
	Src string
	// BitDepth of the output PCM (determines format and codec via RawPCMFormat/RawPCMCodec).
	BitDepth int
	// Channels for the output (-ac). Zero omits -ac, letting ffmpeg preserve the source channel count.
	Channels int
	// Stdout receives the decoded PCM. When nil, output is captured and returned as []byte.
	// Set to io.Discard for benchmarks where the decoded data is not needed.
	Stdout io.Writer
	// Args are optional extra output arguments placed before the output pipe.
	// For example: "-channel_layout", "stereo".
	Args []string
}

// FFmpegDecode decodes an audio file to raw PCM.
// Returns captured PCM bytes when Stdout is nil; returns nil when Stdout is non-nil.
func FFmpegDecode(t *testing.T, opts FFmpegDecodeOptions) []byte {
	t.Helper()

	args := []string{
		"-i", opts.Src,
		"-f", RawPCMFormat(opts.BitDepth),
	}

	if opts.Channels > 0 {
		args = append(args, "-ac", strconv.Itoa(opts.Channels))
	}

	args = append(args, "-acodec", RawPCMCodec(opts.BitDepth))
	args = append(args, opts.Args...)
	args = append(args, "-")

	result := FFmpeg(t, FFmpegOptions{
		Args:   args,
		Stdout: opts.Stdout,
	})

	return result.Stdout
}

// RawPCMFormat returns the ffmpeg raw format name for a given bit depth.
func RawPCMFormat(bitDepth int) string {
	switch bitDepth {
	case BitDepth8:
		return "s8"
	case BitDepth24:
		return "s24le"
	case BitDepth32:
		return "s32le"
	default:
		return "s16le"
	}
}

// RawPCMCodec returns the ffmpeg PCM codec name for a given bit depth.
func RawPCMCodec(bitDepth int) string {
	switch bitDepth {
	case BitDepth8:
		return "pcm_s8"
	case BitDepth24:
		return "pcm_s24le"
	case BitDepth32:
		return "pcm_s32le"
	default:
		return "pcm_s16le"
	}
}
