package agar

import (
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/mod/tigron/test"
)

// ID3Version represents an ID3 tag version.
type ID3Version string

// ID3 tag versions supported by the id3v2 tool.
const (
	ID3v11 ID3Version = "1.1" // ID3v1.1 (limited fields)
	ID3v22 ID3Version = "2.2" // ID3v2.2
	ID3v23 ID3Version = "2.3" // ID3v2.3
	ID3v24 ID3Version = "2.4" // ID3v2.4 (recommended)
)

// id3v2 command flags.
const (
	id3v2FlagV2Only = "--id3v2-only"
	id3v2FlagV2     = "-2"
	id3v2FlagV1     = "-1"
)

// MP3Tags holds metadata for MP3 files.
type MP3Tags struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Year        int
	Track       int
	TrackTotal  int
	Disc        int
	DiscTotal   int
	Genre       string
	Comment     string
	Composer    string
}

// DefaultMP3Tags returns the standard test metadata used across gill tests.
func DefaultMP3Tags() MP3Tags {
	return MP3Tags{
		Title:       "Test Title",
		Artist:      "Test Artist",
		Album:       "Test Album",
		AlbumArtist: "Test AlbumArtist",
		Year:        testYear,
		Track:       testTrack,
		TrackTotal:  testTrackTotal,
		Disc:        testDisc,
		DiscTotal:   0,
		Genre:       "Jazz",
		Comment:     "Test Comment",
		Composer:    "Test Composer",
	}
}

// MP3SetID3Tags writes ID3 tags to an MP3 file using id3v2.
// The version parameter controls which ID3 version to write.
func MP3SetID3Tags(helpers test.Helpers, path string, version ID3Version, tags MP3Tags) {
	helpers.T().Helper()

	args := []string{}

	// Set ID3 version
	if version == ID3v11 {
		args = append(args, id3v2FlagV1)
	} else {
		// ID3v2.2, v2.3, v2.4 all use -2 flag
		args = append(args, id3v2FlagV2, id3v2FlagV2Only)
	}

	// Basic tags supported by all versions
	if tags.Title != "" {
		args = append(args, "-t", tags.Title)
	}

	if tags.Artist != "" {
		args = append(args, "-a", tags.Artist)
	}

	if tags.Album != "" {
		args = append(args, "-A", tags.Album)
	}

	if tags.Year > 0 {
		args = append(args, "-y", strconv.Itoa(tags.Year))
	}

	if tags.Genre != "" {
		args = append(args, "-g", tags.Genre)
	}

	if tags.Comment != "" {
		args = append(args, "-c", tags.Comment)
	}

	// Track number - format depends on version
	if tags.Track > 0 {
		trackStr := strconv.Itoa(tags.Track)
		if tags.TrackTotal > 0 && version != ID3v11 {
			trackStr += "/" + strconv.Itoa(tags.TrackTotal)
		}

		args = append(args, "-T", trackStr)
	}

	// ID3v2-only frames
	if version != ID3v11 {
		// Album artist (TPE2)
		if tags.AlbumArtist != "" {
			args = append(args, "--TPE2", tags.AlbumArtist)
		}

		// Composer (TCOM)
		if tags.Composer != "" {
			args = append(args, "--TCOM", tags.Composer)
		}

		// Disc number (TPOS)
		if tags.Disc > 0 {
			discStr := strconv.Itoa(tags.Disc)
			if tags.DiscTotal > 0 {
				discStr += "/" + strconv.Itoa(tags.DiscTotal)
			}

			args = append(args, "--TPOS", discStr)
		}
	}

	args = append(args, path)

	helpers.Custom(id3v2Binary, args...).Run(&test.Expected{})
}

// TaggedMP3 returns path to MP3 with ID3v2.4 tags (default).
func TaggedMP3(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return TaggedMP3WithVersion(data, helpers, ID3v24)
}

// TaggedMP3WithVersion returns path to MP3 with specified ID3 version tags.
func TaggedMP3WithVersion(data test.Data, helpers test.Helpers, version ID3Version) string {
	helpers.T().Helper()

	filename := "tagged-mp3-id3v" + string(version) + ".mp3"
	path := generate(helpers, filepath.Join(data.Temp().Dir(), filename), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + shortDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libmp3lame", "-b:a", "256k",
	})

	MP3SetID3Tags(helpers, path, version, DefaultMP3Tags())

	return path
}

// TaggedMP3ID3v24 returns path to MP3 with ID3v2.4 tags.
func TaggedMP3ID3v24(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return TaggedMP3WithVersion(data, helpers, ID3v24)
}

// TaggedMP3ID3v23 returns path to MP3 with ID3v2.3 tags.
func TaggedMP3ID3v23(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return TaggedMP3WithVersion(data, helpers, ID3v23)
}

// TaggedMP3ID3v22 returns path to MP3 with ID3v2.2 tags.
func TaggedMP3ID3v22(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return TaggedMP3WithVersion(data, helpers, ID3v22)
}

// TaggedMP3ID3v11 returns path to MP3 with ID3v1.1 tags.
// Note: ID3v1.1 has limited metadata support (no album artist, composer, disc).
func TaggedMP3ID3v11(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return TaggedMP3WithVersion(data, helpers, ID3v11)
}

// UntaggedMP3 returns path to MP3 with no tags.
func UntaggedMP3(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "untagged.mp3"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + shortDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libmp3lame", "-b:a", "256k",
	})
}
