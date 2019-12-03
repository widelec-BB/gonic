package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"senan.xyz/g/gonic/model"
)

var (
	baseURL = "https://ws.audioscrobbler.com/2.0/"
	client  = &http.Client{
		Timeout: 10 * time.Second,
	}
)

type BaseAuthOptions struct {
	APIKey string
	Secret string
}

type ScrobbleOptions struct {
	BaseAuthOptions
	Session    string
	Track      *model.Track
	StampMili  int
	Submission bool
}

func GetSession(opts BaseAuthOptions, token string) (string, error) {
	params := url.Values{}
	params.Add("method", "auth.getSession")
	params.Add("api_key", opts.APIKey)
	params.Add("token", token)
	params.Add("api_sig", getParamSignature(params, opts.Secret))
	resp, err := makeRequest("GET", params)
	if err != nil {
		return "", errors.Wrap(err, "making session GET")
	}
	return resp.Session.Key, nil
}

func Scrobble(opts ScrobbleOptions) error {
	params := url.Values{}
	if opts.Submission {
		params.Add("method", "track.Scrobble")
		// last.fm wants the timestamp in seconds
		params.Add("timestamp", strconv.Itoa(opts.StampMili/1e3))
	} else {
		params.Add("method", "track.updateNowPlaying")
	}
	params.Add("api_key", opts.APIKey)
	params.Add("sk", opts.Session)
	params.Add("artist", opts.Track.TagTrackArtist)
	params.Add("track", opts.Track.TagTitle)
	params.Add("trackNumber", strconv.Itoa(opts.Track.TagTrackNumber))
	params.Add("album", opts.Track.Album.TagTitle)
	params.Add("mbid", opts.Track.Album.TagBrainzID)
	params.Add("albumArtist", opts.Track.Artist.Name)
	params.Add("api_sig", getParamSignature(params, opts.Secret))
	_, err := makeRequest("POST", params)
	return err
}

func getParamSignature(params url.Values, secret string) string {
	// the parameters must be in order before hashing
	paramKeys := make([]string, 0)
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)
	toHash := ""
	for _, k := range paramKeys {
		toHash += k
		toHash += params[k][0]
	}
	toHash += secret
	hash := md5.Sum([]byte(toHash))
	return hex.EncodeToString(hash[:])
}

func makeRequest(method string, params url.Values) (*LastFM, error) {
	req, _ := http.NewRequest(method, baseURL, nil)
	req.URL.RawQuery = params.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "get")
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	lastfm := &LastFM{}
	err = decoder.Decode(lastfm)
	if err != nil {
		return nil, errors.Wrap(err, "decoding")
	}
	if lastfm.Error != nil {
		return nil, fmt.Errorf("parsing: %v", lastfm.Error.Value)
	}
	return lastfm, nil
}
