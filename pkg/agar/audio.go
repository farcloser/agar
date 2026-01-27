package agar

import (
	"path/filepath"

	"github.com/containerd/nerdctl/mod/tigron/test"
)

const (
	defaultDuration     = "10"
	shortDuration       = "3"
	ffmpegBinary        = "ffmpeg"
	ffprobeBinary       = "ffprobe"
	soxBinary           = "sox"
	metaflacBinary      = "metaflac"
	atomicParsleyBinary = "atomicparsley"
	id3v2Binary         = "id3v2"
	vorbiscommentBinary = "vorbiscomment"
	opustagsBinary      = "opustags"

	// Test metadata constants for consistent test data across formats.
	testYear       = 2000
	testTrack      = 3
	testTrackTotal = 6
	testDisc       = 2
)

// Genuine16bit44k returns path to genuine 16-bit 44.1kHz stereo FLAC.
func Genuine16bit44k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "genuine-16bit-44k.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.5",
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// Genuine24bit96k returns path to genuine 24-bit 96kHz stereo FLAC with ultrasonic content.
func Genuine24bit96k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "genuine-24bit-96k.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.3",
		"-f", "lavfi", "-i", "sine=frequency=25000:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=30000:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=35000:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=40000:duration=" + defaultDuration,
		"-filter_complex", "[0][1][2][3][4]amix=inputs=5:duration=first,pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "96000", "-sample_fmt", "s32",
	})
}

// Genuine24bit48k returns path to genuine 24-bit 48kHz stereo FLAC.
func Genuine24bit48k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "genuine-24bit-48k.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.5",
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "48000", "-sample_fmt", "s32",
	})
}

// GenuineMono16bit44k returns path to genuine mono 16-bit 44.1kHz FLAC.
func GenuineMono16bit44k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "genuine-mono-16bit-44k.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.5",
		"-af", "volume=-6dB",
		"-ac", "1", "-ar", "44100", "-sample_fmt", "s16",
	})
}

// FakeHiresPadded24bit returns path to fake hi-res (16-bit padded to 24-bit container).
func FakeHiresPadded24bit(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generateWithPipe(helpers, filepath.Join(data.Temp().Dir(), "fake-hires-padded-24bit.flac"),
		[]string{
			"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
			"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
			"-ar", "44100", "-sample_fmt", "s16",
			"-f", "wav", "-",
		},
		[]string{
			"-c:a", "flac", "-sample_fmt", "s32",
		},
	)
}

// Upsampled44kTo96k returns path to upsampled file (44.1kHz to 96kHz).
func Upsampled44kTo96k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generateWithPipe(helpers, filepath.Join(data.Temp().Dir(), "upsampled-44k-to-96k.flac"),
		[]string{
			"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.5",
			"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
			"-ar", "44100", "-sample_fmt", "s16",
			"-f", "wav", "-",
		},
		[]string{
			"-af", "aresample=96000", "-sample_fmt", "s32",
		},
	)
}

// FakeStereoMonoDuplicate returns path to fake stereo (mono duplicated to both channels).
func FakeStereoMonoDuplicate(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "fake-stereo-mono-duplicate.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// TrueStereoDifferentChannels returns path to true stereo (different content in L/R).
func TrueStereoDifferentChannels(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "true-stereo-different-channels.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// PhaseCancellationInverted returns path to phase-inverted stereo file.
func PhaseCancellationInverted(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "phase-cancellation-inverted.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=-1*c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// ClippedHard returns path to hard-clipped audio file.
func ClippedHard(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "clipped-hard.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=20dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// ClippedLimited returns path to soft-clipped (limited) audio file.
func ClippedLimited(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "clipped-limited.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=15dB,alimiter=limit=1:attack=0.1:release=10",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DCOffsetPositive returns path to audio with positive DC offset.
func DCOffsetPositive(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dc-offset-positive.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,dcshift=0.1,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DCOffsetNegative returns path to audio with negative DC offset.
func DCOffsetNegative(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dc-offset-negative.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,dcshift=-0.15,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// SilenceMiddleGap returns path to audio with silence gap in the middle.
func SilenceMiddleGap(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "silence-middle-gap.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo:d=6",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-filter_complex", "[0]pan=stereo|c0=c0|c1=c0[a];[2]pan=stereo|c0=c0|c1=c0[b];[a][1][b]concat=n=3:v=0:a=1,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// SilenceLongIntro returns path to audio with long silence at the start.
func SilenceLongIntro(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "silence-long-intro.flac"), []string{
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo:d=5",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=5",
		"-filter_complex", "[1]pan=stereo|c0=c0|c1=c0[a];[0][a]concat=n=2:v=0:a=1,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// TruncatedAbruptCut returns path to audio with abrupt cut (no fade).
func TruncatedAbruptCut(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "truncated-abrupt-cut.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-t", "5.123",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// ProperFadeout returns path to audio with proper fadeout.
func ProperFadeout(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "proper-fadeout.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB,afade=t=out:st=8:d=2",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DynamicsExcellent returns path to audio with excellent dynamics (LRA ~15+ LU).
func DynamicsExcellent(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dynamics-excellent.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.3",
		"-af", "pan=stereo|c0=c0|c1=c0,tremolo=f=0.5:d=0.8,volume=-12dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DynamicsOK returns path to audio with OK dynamics (LRA ~8-10 LU).
func DynamicsOK(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dynamics-ok.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.3",
		"-af", "pan=stereo|c0=c0|c1=c0,tremolo=f=0.5:d=0.8,acompressor=threshold=-20dB:ratio=3:attack=20:release=200,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DynamicsMediocre returns path to audio with mediocre dynamics (LRA ~4-6 LU).
func DynamicsMediocre(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dynamics-mediocre.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.3",
		"-af", "pan=stereo|c0=c0|c1=c0,tremolo=f=0.5:d=0.8,acompressor=threshold=-15dB:ratio=8:attack=5:release=50,alimiter=limit=0.9:attack=1:release=10,volume=3dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// DynamicsFucked returns path to brickwalled audio (LRA ~2 LU).
func DynamicsFucked(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "dynamics-fucked.flac"), []string{
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.3",
		"-af", "pan=stereo|c0=c0|c1=c0,tremolo=f=0.5:d=0.8,acompressor=threshold=-10dB:ratio=20:attack=1:release=10,alimiter=limit=0.95:attack=0.1:release=1,volume=6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// LowLoudnessQuiet returns path to very quiet audio (-30dB).
func LowLoudnessQuiet(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "low-loudness-quiet.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-30dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// MultiStream3Audio returns path to MKV with 3 audio streams.
func MultiStream3Audio(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "multi-stream-3-audio.mkv"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=880:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "anoisesrc=d=" + defaultDuration + ":c=pink:a=0.2",
		"-filter_complex", "[0]pan=stereo|c0=c0|c1=c0,volume=-6dB[a0];[1]pan=stereo|c0=c0|c1=c0,volume=-6dB[a1];[2]pan=stereo|c0=c0|c1=c0,volume=-12dB[a2]",
		"-map", "[a0]", "-map", "[a1]", "-map", "[a2]",
		"-c:a:0", "flac", "-c:a:1", "flac", "-c:a:2", "flac",
		"-metadata:s:a:0", "title=440Hz Sine",
		"-metadata:s:a:1", "title=880Hz Sine",
		"-metadata:s:a:2", "title=Pink Noise",
	})
}

// FormatFLAC returns path to FLAC format test file.
func FormatFLAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-flac.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "flac",
	})
}

// FormatALAC returns path to ALAC format test file.
func FormatALAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-alac.m4a"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "alac",
	})
}

// FormatAAC256k returns path to AAC 256k format test file.
func FormatAAC256k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-aac-256k.m4a"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "aac", "-b:a", "256k",
	})
}

// FormatAAC64k returns path to AAC 64k format test file.
func FormatAAC64k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-aac-64k.m4a"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "aac", "-b:a", "64k",
	})
}

// FormatMP3320k returns path to MP3 320k format test file.
func FormatMP3320k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-mp3-320k.mp3"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libmp3lame", "-b:a", "320k",
	})
}

// FormatMP396k returns path to MP3 96k format test file.
func FormatMP396k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-mp3-96k.mp3"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libmp3lame", "-b:a", "96k",
	})
}

// FormatOggVorbis returns path to OGG Vorbis format test file.
func FormatOggVorbis(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-ogg-vorbis.ogg"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libvorbis", "-q:a", "6",
	})
}

// FormatOpus192k returns path to Opus 192k format test file.
func FormatOpus192k(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-opus-192k.opus"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + defaultDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + defaultDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "48000", "-c:a", "libopus", "-b:a", "192k",
	})
}

// FormatMP4VideoOnly returns path to MP4 with video only (no audio stream).
func FormatMP4VideoOnly(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-mp4-video-only.mp4"), []string{
		"-f", "lavfi", "-i", "testsrc=duration=" + shortDuration + ":size=320x240:rate=30",
		"-c:v", "libx264", "-preset", "ultrafast",
		"-an",
	})
}

// FormatMP4MultiAudio returns path to MP4 with multiple audio streams.
func FormatMP4MultiAudio(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "format-mp4-multi-audio.mp4"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=880:duration=" + shortDuration,
		"-filter_complex", "[0]pan=stereo|c0=c0|c1=c0,volume=-6dB[a0];[1]pan=stereo|c0=c0|c1=c0,volume=-6dB[a1]",
		"-map", "[a0]", "-map", "[a1]",
		"-c:a:0", "aac", "-b:a:0", "128k",
		"-c:a:1", "aac", "-b:a:1", "128k",
	})
}
