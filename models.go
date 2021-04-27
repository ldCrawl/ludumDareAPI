package main

import (
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	Jam   SubmissionType = "jam"
	Compo SubmissionType = "compo"
)

const (
	PlatformMicrosoftWindows Platform = 42337
)

const (
	NodeTypeItem NodeType = "item"
)

const (
	NodeSubTypeGame NodeSubType = "game"
)

type (
	Platform int

	SubmissionType string
	NodeType       string
	NodeSubType    string

	Game struct {
		ID   int
		Name string
		Path string
		Body string

		SubmissionType SubmissionType

		Meta GameMeta
		// Event Event
	}

	Author struct {
		ID     int
		Name   string
		Path   string
		Body   string
		Avatar string
	}

	GameMeta struct {
		Authors   []Author
		Published time.Time
		Created   time.Time
		Modified  time.Time

		Links []GameLink
	}

	Event struct {
		Name  string
		Slug  string
		Body  string
		Theme string

		Grades []Grade

		Start time.Time
		End   time.Time

		Published time.Time
		Created   time.Time
		Modified  time.Time
	}

	Grade struct {
		Name     string
		Optional bool
	}

	GameLink struct {
		Name string
		Link string
		Tags []int
	}
)

type (
	gameMeta struct {
		Authors []int
		Cover   string
		Links   []gameMetaLink
	}

	gameMetaLink struct {
		Link string
		Tags []int
		Name string
	}

	authorMeta struct {
		Avatar string `json:"avatar,omitempty"`
	}
)

func (meta *authorMeta) UnmarshalJSON(data []byte) error {
	const (
		avatar = "avatar"
	)

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		var jsErr *json.UnmarshalTypeError
		if !errors.As(err, &jsErr) {
			return err
		}

		return nil
	}

	if val, ok := v[avatar]; ok {
		meta.Avatar = val.(string)
	}

	return nil
}

func (meta *gameMeta) UnmarshalJSON(data []byte) error {
	const (
		author string = "author"
		cover  string = "cover"

		// \/ link prefix
		link string = "link"
		// \/ link suffixes
		tag  string = "tag"
		name string = "name"
	)

	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if val, ok := v[author]; ok {
		authors := val.([]interface{})
		meta.Authors = make([]int, len(authors))
		for i, author := range authors {
			meta.Authors[i] = int(author.(float64))
		}
	}

	if val, ok := v[cover]; ok {
		meta.Cover = val.(string)
	}

	links := make([]gameMetaLink, 0)
	var lastID int
	var currentLink gameMetaLink
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if !strings.HasPrefix(key, link) {
			continue
		}

		id, err := strconv.Atoi(strings.Split(key, "-")[1])
		if err != nil {
			return err
		}

		if id != lastID {
			if lastID > 0 {
				links = append(links, currentLink)
			}

			currentLink = gameMetaLink{}
			lastID = id
		}

		val := v[key]
		if strings.HasSuffix(key, tag) {
			tags := val.([]interface{})
			currentLink.Tags = make([]int, len(tags))
			for i, tag := range tags {
				currentLink.Tags[i] = int(tag.(float64))
			}
			continue
		}

		if strings.HasSuffix(key, name) {
			currentLink.Name = val.(string)
			continue
		}

		currentLink.Link = val.(string)
	}

	links = append(links, currentLink)
	meta.Links = links

	return nil
}
