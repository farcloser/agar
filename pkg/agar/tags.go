package agar

import (
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/mod/tigron/test"
)

// FLACTags holds metadata for FLAC files using Vorbis comment field names.
type FLACTags struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Date        string
	TrackNumber int
	TrackTotal  int
	DiscNumber  int
	DiscTotal   int
	Genre       string
	Comment     string
	Composer    string
}

// DefaultFLACTags returns the standard test metadata used across gill tests.
func DefaultFLACTags() FLACTags {
	return FLACTags{
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

// AddTag adds a tag to a FLAC file using metaflac.
// This allows adding multiple values for the same tag key.
func AddTag(helpers test.Helpers, path, key, value string) {
	helpers.T().Helper()

	helpers.Custom(metaflacBinary, "--set-tag="+key+"="+value, path).Run(&test.Expected{})
}

// RemoveTag removes all instances of a tag from a FLAC file using metaflac.
func RemoveTag(helpers test.Helpers, path, key string) {
	helpers.T().Helper()

	helpers.Custom(metaflacBinary, "--remove-tag="+key, path).Run(&test.Expected{})
}

// SetTag sets a tag value, removing any existing values for that key first.
func SetTag(helpers test.Helpers, path, key, value string) {
	helpers.T().Helper()

	RemoveTag(helpers, path, key)
	AddTag(helpers, path, key, value)
}

// FLACSetTags writes all tags to a FLAC file using metaflac.
func FLACSetTags(helpers test.Helpers, path string, tags FLACTags) {
	helpers.T().Helper()

	if tags.Title != "" {
		SetTag(helpers, path, "TITLE", tags.Title)
	}

	if tags.Artist != "" {
		SetTag(helpers, path, "ARTIST", tags.Artist)
	}

	if tags.Album != "" {
		SetTag(helpers, path, "ALBUM", tags.Album)
	}

	if tags.AlbumArtist != "" {
		SetTag(helpers, path, "ALBUMARTIST", tags.AlbumArtist)
	}

	if tags.Date != "" {
		SetTag(helpers, path, "DATE", tags.Date)
	}

	if tags.Genre != "" {
		SetTag(helpers, path, "GENRE", tags.Genre)
	}

	if tags.Comment != "" {
		SetTag(helpers, path, "COMMENT", tags.Comment)
	}

	if tags.Composer != "" {
		SetTag(helpers, path, "COMPOSER", tags.Composer)
	}

	if tags.TrackNumber > 0 {
		SetTag(helpers, path, "TRACKNUMBER", strconv.Itoa(tags.TrackNumber))
	}

	if tags.TrackTotal > 0 {
		SetTag(helpers, path, "TRACKTOTAL", strconv.Itoa(tags.TrackTotal))
	}

	if tags.DiscNumber > 0 {
		SetTag(helpers, path, "DISCNUMBER", strconv.Itoa(tags.DiscNumber))
	}

	if tags.DiscTotal > 0 {
		SetTag(helpers, path, "DISCTOTAL", strconv.Itoa(tags.DiscTotal))
	}
}

// TaggedFLAC returns path to FLAC with standard metadata tags.
func TaggedFLAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := generate(helpers, filepath.Join(data.Temp().Dir(), "tagged.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})

	FLACSetTags(helpers, path, DefaultFLACTags())

	return path
}

// UntaggedFLAC returns path to FLAC with no tags.
func UntaggedFLAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	return generate(helpers, filepath.Join(data.Temp().Dir(), "untagged.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
	})
}

// TaggedFLACMultiArtist returns path to FLAC with multiple artist values.
// Uses metaflac to add duplicate ARTIST tags since FFmpeg doesn't support this.
func TaggedFLACMultiArtist(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := generate(helpers, filepath.Join(data.Temp().Dir(), "tagged-multi-artist.flac"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-af", "pan=stereo|c0=c0|c1=c0,volume=-6dB",
		"-ar", "44100", "-sample_fmt", "s16",
		"-metadata", "artist=Artist One",
		"-metadata", "album=Collaboration Album",
	})

	// Add second artist using metaflac
	AddTag(helpers, path, "ARTIST", "Artist Two")

	return path
}
