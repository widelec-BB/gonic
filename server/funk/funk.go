package funk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"

	"senan.xyz/g/gonic/model"
)

// TODO: if funk.pub becomes a real thing
// - not store the funk usernames and passwords, use an
//   api key or something similar
// - not sign in for every request
// - probably not use graphql either, audioscrobbler api
//   should be implemented by then

type FunkOptions struct {
	BaseURL  string
	Username string
	Password string
	Track    *model.Track
}

type SignInResponse struct {
	SignIn struct {
		Token string `json:"token"`
	} `json:"signIn"`
}

type ListenContent struct {
	Artist string `json:"artistName"`
	Track  string `json:"trackName"`
	Album  string `json:"albumName"`
}

type ListenPayload struct {
	ListenContent `json:"content"`
}

func Funk(opts FunkOptions) error {
	baseURL := fmt.Sprintf("%s/graphql", opts.BaseURL)
	client := graphql.NewClient(baseURL)
	req := graphql.NewRequest(`
        mutation ($username: String!, $password: String!) {
            signIn(username: $username, password: $password) {
                token
            }
        }
	`)
	req.Var("username", opts.Username)
	req.Var("password", opts.Password)
	resp := &SignInResponse{}
	if err := client.Run(context.Background(), req, resp); err != nil {
		return errors.Wrap(err, "getting token")
	}
	if resp.SignIn.Token == "" {
		return errors.New("token not returned")
	}
	listen := ListenPayload{ListenContent{
		Artist: opts.Track.TagTrackArtist,
		Track:  opts.Track.TagTitle,
		Album:  opts.Track.Album.TagTitle,
	}}
	listenJSON, err := json.Marshal(listen)
	if err != nil {
		return errors.Wrap(err, "marshalling listen to json")
	}
	req = graphql.NewRequest(`
        mutation ($content: String!) {
            listen(content: $content)
        }
	`)
	req.Var("content", string(listenJSON))
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", resp.SignIn.Token))
	if err := client.Run(context.Background(), req, nil); err != nil {
		return errors.Wrap(err, "sending listen")
	}
	return nil
}
