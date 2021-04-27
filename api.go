package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const (
	walk string = "walk/"
	get  string = "get/"

	events    string = "1/events/"
	ludumDare string = "ludum-dare"
)

type (
	API interface {
		GetGame(name string) (Game, error)
	}

	api struct {
		resty *resty.Client
	}
)

func New() API {
	r := resty.New()
	r.HostURL = "https://api.ldjam.com/vx/node2/"

	return &api{
		resty: r,
	}
}

func (api *api) GetGame(name string) (Game, error) {
	req := api.resty.R()
	// TODO determine how to pass the ld event efficiently withot getting into the way ...
	resp, err := req.Get(walk + events + ludumDare + "/48/" + name)
	if err != nil {
		return Game{}, err
	}

	var walkGameResponse struct {
		NodeID int `json:"node_id"`
	}

	if err := json.Unmarshal(resp.Body(), &walkGameResponse); err != nil {
		return Game{}, err
	}

	var rawGame struct {
		Game [1]struct {
			ID             int            `json:"id,omitempty"`
			Parent         int            `json:"parent,omitempty"`
			SubmissionType SubmissionType `json:"subsubtype,omitempty"`
			Type           NodeType       `json:"type,omitempty"`
			SubType        NodeSubType    `json:"subtype,omitempty"`
			Name           string         `json:"name,omitempty"`
			Body           string         `json:"body,omitempty"`
			Path           string         `json:"path,omitempty"`

			Meta gameMeta `json:"meta,omitempty"`

			// Grade struct{} `json:"grade"`
			Magic struct {
				Cool     float64 `json:"cool,omitempty"`
				Feedback float64 `json:"feedback,omitempty"`
				Given    float64 `json:"given,omitempty"`
				Grade    float64 `json:"grade,omitempty"`
				Smart    float64 `json:"smart,omitempty"`
			} `json:"magic,omitempty"`

			Published time.Time `json:"published,omitempty"`
			Created   time.Time `json:"created,omitempty"`
			Modified  time.Time `json:"modified,omitempty"`
		} `json:"node,omitempty"`
	}

	if err := api.Get(walkGameResponse.NodeID, &rawGame); err != nil {
		return Game{}, nil
	}

	game := &rawGame.Game[0]

	if game.Type != NodeTypeItem || game.SubType != NodeSubTypeGame {
		return Game{}, ErrNotAGame
	}

	authorsCount := len(game.Meta.Authors)
	authors := make([]Author, authorsCount)
	links := make([]GameLink, len(game.Meta.Links))

	outChan := make(chan Author, authorsCount)
	errChan := make(chan error)

	defer close(outChan)
	defer close(errChan)

	go api.retrieveAuthors(game.Meta.Authors, outChan, errChan)

	for i := 0; i < authorsCount; i++ {
		select {
		case author := <-outChan:
			authors[i] = author
		case err := <-errChan:
			return Game{}, err
		}
	}

	for i := range game.Meta.Links {
		link := game.Meta.Links[i]
		links[i] = GameLink{
			Name: link.Name,
			Link: link.Link,
			Tags: link.Tags,
		}
	}

	return Game{
		ID:             game.ID,
		Name:           game.Name,
		SubmissionType: game.SubmissionType,
		Path:           game.Path,
		Body:           game.Body,
		Meta: GameMeta{
			Authors:   authors,
			Published: game.Published,
			Created:   game.Created,
			Modified:  game.Modified,
			Links:     links,
		},
		// Event: Event{},
	}, nil
}

func (api *api) GetAuthors(authors ...int) ([]Author, error) {
	return nil, nil
}

func (api *api) Get(nodeID int, outVal interface{}) error {
	req := api.resty.R()
	resp, err := req.Get(get + strconv.Itoa(nodeID))

	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Body(), outVal); err != nil {
		return err
	}

	return nil
}

func (api *api) retrieveAuthors(authors []int, outChan chan<- Author, errChan chan<- error) {
	for _, id := range authors {
		var rawAuthor struct {
			Author [1]struct {
				ID   int        `json:"id,omitempty"`
				Name string     `json:"name,omitempty"`
				Path string     `json:"path,omitempty"`
				Body string     `json:"body,omitempty"`
				Meta authorMeta `json:"meta,omitempty"`
			} `json:"node,omitempty"`
		}

		if err := api.Get(id, &rawAuthor); err != nil {
			errChan <- errors.Wrapf(err, "could not load user with ID %d", id)
		}

		author := rawAuthor.Author[0]
		if author.ID != id {
			errChan <- errors.New("Sum Ting Wong")
		}

		outChan <- Author{
			ID:     id,
			Name:   author.Name,
			Path:   author.Path,
			Body:   author.Body,
			Avatar: author.Meta.Avatar,
		}
	}
}
