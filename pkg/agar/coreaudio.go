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
	"errors"
	"testing"
	"time"
)

// ErrCoreAudioUnavailable is returned when AudioToolbox is not available
// (non-darwin platform or CGO disabled).
var ErrCoreAudioUnavailable = errors.New("coreaudio: AudioToolbox not available on this platform")

// BenchDecodeCoreAudio benchmarks CoreAudio decoding via CGO (in-process).
// It skips the test if CoreAudio is not available on the current platform.
func BenchDecodeCoreAudio(t *testing.T, format BenchFormat, opts BenchOptions, encoded []byte) BenchResult {
	t.Helper()

	// Verify CGO decode works before benchmarking.
	if _, err := CoreAudioDecode(encoded); err != nil {
		if errors.Is(err, ErrCoreAudioUnavailable) {
			t.Skip("CoreAudio not available on this platform")
		}

		t.Fatalf("coreaudio: %v", err)
	}

	opts = opts.WithDefaults()
	durations := make([]time.Duration, opts.Iterations)

	for iter := range opts.Iterations {
		start := time.Now()

		_, err := CoreAudioDecode(encoded)
		if err != nil {
			t.Fatalf("coreaudio iter %d: %v", iter, err)
		}

		durations[iter] = time.Since(start)
	}

	return ComputeResult(format, "coreaudio", "decode", durations, len(encoded))
}
