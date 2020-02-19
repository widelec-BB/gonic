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
		mutation ($artist: String!, $track: String!, $album: String) {
			listen(artist: $artist, track: $track, album: $album)
		}
	`)
	req.Var("artist", opts.Track.TagTrackArtist)
	req.Var("track", opts.Track.TagTitle)
	req.Var("album", opts.Track.Album.TagTitle)
	req.Header.Set("Authorization", fmt.Sprintf("%s;%s", opts.Username, opts.Password))
	if err := client.Run(context.Background(), req, nil); err != nil {
		return errors.Wrap(err, "sending listen")
	}
	return nil
}
