# Known issues

### Cannot generate audio with dropout artifacts

> **OPEN**

All agar fixtures use ffmpeg's `lavfi` synthesis sources (`anoisesrc`, `sine`,
`anullsrc`) to generate test audio. These sources produce mathematically
perfect, continuous waveforms with no sample-level discontinuities.

Dropout detection in haustorium looks for two types of artifacts:
- **Zero runs**: consecutive zero-valued samples lasting >= 1ms, indicating
  a glitch where the audio stream was interrupted
- **Delta discontinuities**: sample-to-sample jumps where one side is near
  zero, indicating an abrupt transition between audio and silence

Neither artifact can be produced by ffmpeg's synthesis pipeline. Even
concatenating segments with `concat` or applying filters produces clean
transitions at the sample level.

**Impact:** haustorium's dropout integration test (`TestDropoutsPositive`)
is skipped. Only the negative case (clean audio has no dropouts) is tested.

**Resolution:** add a fixture function that writes raw PCM samples directly
using Go's `encoding/binary` package, injecting:
1. A zero run of ~50ms at a known position in an otherwise normal waveform
2. A delta discontinuity (normal sample followed by zero followed by normal
   sample) at a known position

The fixture would write a temporary WAV file with the constructed samples,
then optionally convert to FLAC via ffmpeg. This bypasses ffmpeg's synthesis
entirely for the artifact portion while still producing a valid audio file.

### Cannot generate audio with inter-sample peaks

> **OPEN**

Inter-sample peak (ISP) detection counts events where the reconstructed
analog signal exceeds 0 dBTP (digital full scale). This happens when the
sinc-interpolated waveform between two digital samples overshoots the
maximum sample value.

ffmpeg's synthesis sources and limiters cannot produce audio with true peak
above 0 dBTP. The limiter (`alimiter`) prevents samples from reaching
digital full scale, and even near-Nyquist sine waves at maximum amplitude
produce true peaks well below 0 dBTP:
- Near-Nyquist sine at 0 dBFS: true peak -18.5 dBTP
- Multi-sine high amplitude: true peak -12.0 dBTP
- Limited audio at 0 dBFS: true peak -8.1 dBTP

ISP events require specific inter-sample value patterns. The classic case
is two consecutive samples at or near +1.0 and -1.0 (or vice versa) at
frequencies close to the Nyquist limit, where the sinc interpolation
reconstructs a waveform that swings beyond the sample values. ffmpeg's
sample-level output is determined by its synthesis algorithms and cannot
be controlled at this granularity.

**Impact:** haustorium's ISP integration test (`TestInterSamplePeaks`)
is entirely skipped. No positive detection test exists.

**Resolution:** add a fixture function that writes raw PCM samples with
carefully chosen values that produce ISP events. For 16-bit 44.1 kHz audio,
alternating samples near +32767 and -32767 at certain phase relationships
will cause the sinc interpolation to overshoot 0 dBTP. The exact sample
values can be determined by running haustorium's own true peak analyzer
on candidate patterns and verifying the output exceeds 0 dBTP.

The fixture would write a temporary WAV file with a short segment of
ISP-triggering samples embedded in an otherwise normal waveform.
