package agar

import (
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/mod/tigron/test"
)

// vorbiscomment command flags.
const vorbisTagFlag = "-t"

// OggTags holds metadata for OGG Vorbis files.
// Uses standard Vorbis comment field names.
type OggTags struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Date        string // Vorbis uses DATE, not YEAR
	TrackNumber int
	TrackTotal  int
	DiscNumber  int
	DiscTotal   int
	Genre       string
	Comment     string
	Composer    string
}

// DefaultOggTags returns the standard test metadata used across gill tests.
func DefaultOggTags() OggTags {
	return OggTags{
		Title:       "Test Title",
		Artist:      "Test Artist",
		Album:       "Test Album",
		AlbumArtist: "Test AlbumArtist",
		Date:        strconv.Itoa(testYear),
		TrackNumber: testTrack,
		TrackTotal:  testTrackTotal,
		DiscNumber:  testDisc,
		DiscTotal:   0,
		Genre:       "Jazz",
		Comment:     "Test Comment",
		Composer:    "Test Composer",
	}
}

// OggSetTags writes Vorbis comments to an OGG file using vorbiscomment.
// Replaces all existing comments.
func OggSetTags(helpers test.Helpers, path string, tags OggTags) {
	helpers.T().Helper()

	args := []string{"-w"} // Write mode (replace all)

	if tags.Title != "" {
		args = append(args, vorbisTagFlag, "TITLE="+tags.Title)
	}

	if tags.Artist != "" {
		args = append(args, vorbisTagFlag, "ARTIST="+tags.Artist)
	}

	if tags.Album != "" {
		args = append(args, vorbisTagFlag, "ALBUM="+tags.Album)
	}

	if tags.AlbumArtist != "" {
		args = append(args, vorbisTagFlag, "ALBUMARTIST="+tags.AlbumArtist)
	}

	if tags.Date != "" {
		args = append(args, vorbisTagFlag, "DATE="+tags.Date)
	}

	if tags.Genre != "" {
		args = append(args, vorbisTagFlag, "GENRE="+tags.Genre)
	}

	if tags.Comment != "" {
		args = append(args, vorbisTagFlag, "COMMENT="+tags.Comment)
	}

	if tags.Composer != "" {
		args = append(args, vorbisTagFlag, "COMPOSER="+tags.Composer)
	}

	// Track number
	if tags.TrackNumber > 0 {
		args = append(args, vorbisTagFlag, "TRACKNUMBER="+strconv.Itoa(tags.TrackNumber))
	}

	if tags.TrackTotal > 0 {
		args = append(args, vorbisTagFlag, "TRACKTOTAL="+strconv.Itoa(tags.TrackTotal))
	}

	// Disc number
	if tags.DiscNumber > 0 {
		args = append(args, vorbisTagFlag, "DISCNUMBER="+strconv.Itoa(tags.DiscNumber))
	}

	if tags.DiscTotal > 0 {
		args = append(args, vorbisTagFlag, "DISCTOTAL="+strconv.Itoa(tags.DiscTotal))
	}

	args = append(args, path)

	helpers.Custom(vorbiscommentBinary, args...).Run(&test.Expected{})
}

// OggAddTag adds a single tag to an OGG file using vorbiscomment.
// This allows adding multiple values for the same tag key.
func OggAddTag(helpers test.Helpers, path, key, value string) {
	helpers.T().Helper()

	helpers.Custom(vorbiscommentBinary, "-a", vorbisTagFlag, key+"="+value, path).Run(&test.Expected{})
}

// TaggedOggVorbis returns path to OGG Vorbis with standard metadata tags.
func TaggedOggVorbis(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := generate(helpers, filepath.Join(data.Temp().Dir(), "tagged-ogg-vorbis.ogg"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + shortDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libvorbis", "-q:a", "6",
	})

	OggSetTags(helpers, path, DefaultOggTags())

	return path
}

// TaggedOggVorbisMultiArtist returns path to OGG with multiple artist values.
func TaggedOggVorbisMultiArtist(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := TaggedOggVorbis(data, helpers)

	// Add second artist using vorbiscomment append mode
	OggAddTag(helpers, path, "ARTIST", "Artist Two")

	return path
}

// UntaggedOggVorbis returns path to OGG Vorbis with no tags.
func UntaggedOggVorbis(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "untagged-ogg-vorbis.ogg"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + shortDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "libvorbis", "-q:a", "6",
	})
}
