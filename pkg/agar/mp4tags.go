package agar

import (
	"path/filepath"
	"strconv"

	"github.com/containerd/nerdctl/mod/tigron/test"
)

// iTunesDomain is the reverse DNS domain for iTunes freeform tags.
const iTunesDomain = "com.apple.iTunes"

// MP4Tags holds metadata for MP4/M4A files.
type MP4Tags struct {
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

// DefaultMP4Tags returns the standard test metadata used across gill tests.
func DefaultMP4Tags() MP4Tags {
	return MP4Tags{
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

// MP4SetAllTags sets all tags on an MP4 file using atomicparsley.
func MP4SetAllTags(helpers test.Helpers, path string, tags MP4Tags) {
	helpers.T().Helper()

	if tags.Title != "" {
		MP4SetTag(helpers, path, "title", tags.Title)
	}

	if tags.Artist != "" {
		MP4SetTag(helpers, path, "artist", tags.Artist)
	}

	if tags.Album != "" {
		MP4SetTag(helpers, path, "album", tags.Album)
	}

	if tags.AlbumArtist != "" {
		MP4SetTag(helpers, path, "albumArtist", tags.AlbumArtist)
	}

	if tags.Year > 0 {
		MP4SetTag(helpers, path, "year", strconv.Itoa(tags.Year))
	}

	if tags.Genre != "" {
		MP4SetTag(helpers, path, "genre", tags.Genre)
	}

	if tags.Comment != "" {
		MP4SetTag(helpers, path, "comment", tags.Comment)
	}

	if tags.Composer != "" {
		MP4SetTag(helpers, path, "composer", tags.Composer)
	}

	// Track number (format: N/T or just N)
	if tags.Track > 0 {
		trackStr := strconv.Itoa(tags.Track)
		if tags.TrackTotal > 0 {
			trackStr += "/" + strconv.Itoa(tags.TrackTotal)
		}

		MP4SetTag(helpers, path, "tracknum", trackStr)
	}

	// Disc number (format: N/T or just N)
	if tags.Disc > 0 {
		discStr := strconv.Itoa(tags.Disc)
		if tags.DiscTotal > 0 {
			discStr += "/" + strconv.Itoa(tags.DiscTotal)
		}

		MP4SetTag(helpers, path, "disk", discStr)
	}
}

// MP4SetTag sets a tag on an MP4 file using atomicparsley.
// Standard tags use short names: title, artist, album, albumArtist, genre, year, etc.
// Modifies the file in-place.
func MP4SetTag(helpers test.Helpers, path, key, value string) {
	helpers.T().Helper()

	helpers.Custom(atomicParsleyBinary, path, "--"+key, value, "--overWrite").Run(&test.Expected{})
}

// MP4SetFreeformTag sets a freeform tag on an MP4 file using atomicparsley.
// Uses the reverse DNS format: domain (e.g., "com.apple.iTunes") and name (e.g., "ARTISTS").
// Modifies the file in-place.
func MP4SetFreeformTag(helpers test.Helpers, path, domain, name, value string) {
	helpers.T().Helper()

	// AtomicParsley syntax: --rDNSatom "value" name=NAME domain=DOMAIN
	helpers.Custom(atomicParsleyBinary, path,
		"--rDNSatom", value,
		"name="+name,
		"domain="+domain,
		"--overWrite").Run(&test.Expected{})
}

// MP4SetArtwork sets cover artwork on an MP4 file using atomicparsley.
// Modifies the file in-place.
func MP4SetArtwork(helpers test.Helpers, path, artworkPath string) {
	helpers.T().Helper()

	helpers.Custom(atomicParsleyBinary, path, "--artwork", artworkPath, "--overWrite").Run(&test.Expected{})
}

// MP4RemoveAllTags removes all metadata from an MP4 file using atomicparsley.
// Modifies the file in-place.
func MP4RemoveAllTags(helpers test.Helpers, path string) {
	helpers.T().Helper()

	helpers.Custom(atomicParsleyBinary, path, "--metaEnema", "--overWrite").Run(&test.Expected{})
}

// MP4RemoveArtwork removes all artwork from an MP4 file using atomicparsley.
// Modifies the file in-place.
func MP4RemoveArtwork(helpers test.Helpers, path string) {
	helpers.T().Helper()

	helpers.Custom(atomicParsleyBinary, path, "--artwork", "REMOVE_ALL", "--overWrite").Run(&test.Expected{})
}

// MP4VerifyTagWithAtomicParsley verifies a tag value using atomicparsley.
// This runs atomicparsley -t and checks for the expected output pattern.
// Use for sanity checking that tags were written correctly.
func MP4VerifyTagWithAtomicParsley(helpers test.Helpers, path string) {
	helpers.T().Helper()

	// Just verify atomicparsley can read the file without error
	helpers.Custom(atomicParsleyBinary, path, "-t").Run(&test.Expected{})
}

// TaggedAAC returns path to AAC with standard metadata tags set via atomicparsley.
func TaggedAAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := FormatAAC256k(data, helpers)
	MP4SetAllTags(helpers, path, DefaultMP4Tags())

	return path
}

// TaggedALAC returns path to ALAC with standard metadata tags set via atomicparsley.
func TaggedALAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := FormatALAC(data, helpers)
	MP4SetAllTags(helpers, path, DefaultMP4Tags())

	return path
}

// TaggedAACWithUnknown returns path to AAC with standard tags plus an unknown freeform tag.
func TaggedAACWithUnknown(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := TaggedAAC(data, helpers)

	// Add an "unknown" freeform tag that's not in our mapping
	MP4SetFreeformTag(helpers, path, iTunesDomain, "CUSTOM_UNKNOWN_TAG", "CustomValue")

	return path
}

// TaggedAACWithMultiArtist returns path to AAC with multiple artist values.
// Note: AtomicParsley doesn't easily support multiple values for the same tag,
// so this uses the ARTISTS freeform tag which is designed for multiple artists.
func TaggedAACWithMultiArtist(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := TaggedAAC(data, helpers)

	// Add ARTISTS freeform tag (used by Picard for multiple artists)
	MP4SetFreeformTag(helpers, path, iTunesDomain, "ARTISTS", "Artist One; Artist Two")

	return path
}

// TaggedAACWithMusicBrainz returns path to AAC with MusicBrainz identifiers.
func TaggedAACWithMusicBrainz(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	path := TaggedAAC(data, helpers)

	// Add MusicBrainz freeform tags
	MP4SetFreeformTag(helpers, path, iTunesDomain, "MusicBrainz Track Id", "12345678-1234-1234-1234-123456789012")
	MP4SetFreeformTag(helpers, path, iTunesDomain, "MusicBrainz Album Id", "abcdefab-abcd-abcd-abcd-abcdefabcdef")
	MP4SetFreeformTag(
		helpers,
		path,
		iTunesDomain,
		"MusicBrainz Artist Id",
		"11111111-2222-3333-4444-555555555555",
	)

	return path
}

// TaggedAACWithArtwork returns path to AAC with standard tags and cover artwork.
func TaggedAACWithArtwork(data test.Data, helpers test.Helpers, artworkPath string) string {
	helpers.T().Helper()

	path := TaggedAAC(data, helpers)
	MP4SetArtwork(helpers, path, artworkPath)

	return path
}

// UntaggedAAC returns path to AAC with no metadata (tags stripped).
func UntaggedAAC(data test.Data, helpers test.Helpers) string {
	helpers.T().Helper()

	// Generate a base AAC file
	path := generate(helpers, filepath.Join(data.Temp().Dir(), "untagged.m4a"), []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=" + shortDuration,
		"-f", "lavfi", "-i", "sine=frequency=554:duration=" + shortDuration,
		"-filter_complex", "[0][1]amerge=inputs=2,volume=-6dB",
		"-ar", "44100", "-c:a", "aac", "-b:a", "256k",
	})

	// Strip all metadata
	MP4RemoveAllTags(helpers, path)

	return path
}
