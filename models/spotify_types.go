package models

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SpotifyUser struct {
	DisplayName  string      `json:"display_name,omitempty"`
	ExternalUrls ExternalUrl `json:"external_urls"`
	Followers    Followers   `json:"followers"`
	Href         string      `json:"href"`
	Id           string      `json:"id"`
	Images       []Image     `json:"images"`
	Type         string      `json:"type"`
	Uri          string      `json:"uri"`
}

type Followers struct {
	Href  string `json:"href,omitempty"`
	Total int    `json:"total"`
}

type LikedTracksResult struct {
	Href     string                  `json:"href"`
	Limit    int                     `json:"limit"`
	Next     string                  `json:"next"`
	Offset   int                     `json:"offset"`
	Previous string                  `json:"previous"`
	Total    int                     `json:"total"`
	Items    []LikedTracksResultItem `json:"items"`
}

type LikedTracksResultItem struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

type Track struct {
	Album            Album        `json:"album"`
	Artists          []Artist     `json:"artists"`
	AvailableMarkets []string     `json:"available_markets"`
	DiscNumber       int          `json:"disc_number"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalIds      ExternalId   `json:"external_ids"`
	ExternalUrls     ExternalUrl  `json:"external_urls"`
	Href             string       `json:"href"`
	Id               string       `json:"id"`
	IsPlayable       bool         `json:"is_playable"`
	Restriction      Restrictions `json:"restrictions"`
	Name             string       `json:"name"`
	Popularity       int          `json:"popularity"`
	TrackNumber      int          `json:"track_number"`
	Type             string       `json:"type"`
	Uri              string       `json:"uri"`
	IsLocal          bool         `json:"is_local"`
}

type Album struct {
	AlbumType            string       `json:"album_type"`
	TotalTracks          int          `json:"total_tracks"`
	AvailableMarkets     []string     `json:"available_markets"`
	ExternalUrls         ExternalUrl  `json:"external_urls"`
	Href                 string       `json:"href"`
	Id                   string       `json:"id"`
	Images               []Image      `json:"images"`
	Name                 string       `json:"name"`
	ReleaseDate          string       `json:"release_date"`
	ReleaseDatePrecision string       `json:"release_date_precision"`
	Restrictions         Restrictions `json:"resctictions,omitempty"`
	Type                 string       `json:"type"`
	Uri                  string       `json:"uri"`
	Artists              []Artist     `json:"artists"`
}

type Artist struct {
	Href         string      `json:"href"`
	Id           string      `json:"id"`
	Name         string      `json:"name"`
	Uri          string      `json:"uri"`
	Type         string      `json:"type"`
	ExternalUrls ExternalUrl `json:"external_urls"`
}

type Image struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ExternalId struct {
	Isrc string `json:"isrc"`
	Ean  string `json:"ean"`
	Upc  string `json:"upc"`
}

type ExternalUrl struct {
	Spotify string `json:"spotify"`
}

type OEmbed struct {
	Html            string `json:"html"`
	Height          int    `json:"height"`
	Version         string `json:"version"`
	ProviderName    string `json:"provider_name"`
	ProvideerURL    string `json:"provider_url"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	ThumbnailURL    string `json:"thumbnail_url"`
	ThumbnailWidth  int    `json:"thumbnail_width"`
	ThumbnailHeight int    `json:"thumbnail_height"`
}

func (e *ExternalUrl) getEmbed() (OEmbed, error) {
	url := fmt.Sprintf("https://open.spotify.com/oembed?url=%s", e.Spotify)
	resp, err := http.Get(url)
	if err != nil {
		return OEmbed{}, err
	}
	defer resp.Body.Close()
	var oembed *OEmbed
	if err := json.NewDecoder(resp.Body).Decode(&oembed); err != nil {
		return OEmbed{}, err
	}
	return *oembed, nil
}

type Restrictions struct {
	Reason string `json:"reason"`
}
