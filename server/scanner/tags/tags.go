package tags

import (
	"strconv"
	"strings"

	"github.com/nicksellen/audiotags"
)

func intSep(in, sep string) int {
	if in == "" {
		return 0
	}
	start := strings.SplitN(in, sep, 2)[0]
	out, err := strconv.Atoi(start)
	if err != nil {
		return 0
	}
	return out
}

type Tags struct {
	raw    map[string]string
	props  *audiotags.AudioProperties
	images map[string][]byte
}

func New(path string) (*Tags, error) {
	raw, props, images, err := audiotags.Read(path)
	return &Tags{raw, props, images}, err
}

func (t *Tags) firstTag(keys ...string) string {
	for _, key := range keys {
		if val, ok := t.raw[key]; ok {
			return val
		}
	}
	return ""
}

func (t *Tags) firstImage(key string) []byte {
	for id, image := range t.images {
		if strings.HasPrefix(id, key) {
			return image
		}
	}
	return nil
}

func (t *Tags) Title() string         { return t.firstTag("title") }
func (t *Tags) BrainzID() string      { return t.firstTag("musicbrainz_trackid") }
func (t *Tags) Artist() string        { return t.firstTag("artist") }
func (t *Tags) Album() string         { return t.firstTag("album") }
func (t *Tags) AlbumArtist() string   { return t.firstTag("albumartist", "album artist") }
func (t *Tags) AlbumBrainzID() string { return t.firstTag("musicbrainz_albumid") }
func (t *Tags) Genre() string         { return t.firstTag("genre") }
func (t *Tags) Year() int             { return intSep(t.firstTag("date", "year"), "-") } // eg. 2019-6-11
func (t *Tags) TrackNumber() int      { return intSep(t.firstTag("tracknumber"), "/") }  // eg. 5/12
func (t *Tags) DiscNumber() int       { return intSep(t.firstTag("discnumber"), "/") }   // eg. 1/2
func (t *Tags) Length() int           { return t.props.Length }
func (t *Tags) Bitrate() int          { return t.props.Bitrate }
func (t *Tags) Cover() []byte         { return t.firstImage("cover") }
