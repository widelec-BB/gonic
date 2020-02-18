package funk

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"senan.xyz/g/gonic/model"
)

type FunkOptions struct {
	BaseURL  string
	Username string
	Password string
	Track    *model.Track
}

func Funk(opts FunkOptions) error {
	baseURL := fmt.Sprintf("%s/graphql", opts.BaseURL)
	client := graphql.NewClient(baseURL)
	req := graphql.NewRequest(`
		mutation ($artist: String!, $track: String!, $album: String, $albumArt: String) {
			listen(artist: $artist, track: $track, album: $album, albumArt: $albumArt)
		}
	`)
	req.Var("artist", opts.Track.TagTrackArtist)
	req.Var("track", opts.Track.TagTitle)
	req.Var("album", opts.Track.Album.TagTitle)
	req.Var("albumArt", fmt.Sprintf("https://gonic.home.senan.xyz/rest/getCoverArt.view?id=%d", opts.Track.Album.ID))
	req.Header.Set("Authorization", fmt.Sprintf("%s;%s", opts.Username, opts.Password))
	if err := client.Run(context.Background(), req, nil); err != nil {
		return errors.Wrap(err, "sending listen")
	}
	return nil
}
