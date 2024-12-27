package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type SpotifyToken struct {
	gorm.Model
	UserID       uint
	User         User
	ID           uint   `gorm:"primaryKey"`
	AccessToken  string `gorm:"not null"`
	RefreshToken string `gorm:"not null"`
}

func (t SpotifyToken) getOauth2Token() (oauth2.Token, error) {
	return oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}, nil
}

func (t SpotifyToken) getRandomLikedTrack(c context.Context, oauth2Conf *oauth2.Config) (LikedTracksResultItem, error) {
	token, err := t.getOauth2Token()
	if err != nil {
		fmt.Println("Error getting oauth2 token")
		return LikedTracksResultItem{}, err
	}

	tokenSource := oauth2Conf.TokenSource(c, &token)
	client := oauth2.NewClient(c, tokenSource)

	// call spotify API
	resp, err := client.Get("https://api.spotify.com/v1/me/tracks?limit=1&offset=0")
	if err != nil {
		fmt.Println("Error calling spotify API")
		return LikedTracksResultItem{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		return LikedTracksResultItem{}, err
	}
	var likedTracks LikedTracksResult
	fmt.Println(string(body))
	if err := json.Unmarshal(body, &likedTracks); err != nil {
		fmt.Println(string(body))
		fmt.Println("Error unmarshalling liked tracks")
		return LikedTracksResultItem{}, err
	}
	RandSource := DefaultRandSource{}
	track_n := RandSource.Intn(likedTracks.Total)

	fmt.Println("track_n: ", track_n)
	// Identify n-th track in the list
	url := fmt.Sprintf("https://api.spotify.com/v1/me/tracks?limit=1&offset=%d", track_n)
	resp, err = client.Get(url)
	if err != nil {
		fmt.Println("Error calling Spotify API")
		return LikedTracksResultItem{}, err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		return LikedTracksResultItem{}, err
	}
	if err := json.Unmarshal(body, &likedTracks); err != nil {
		fmt.Println("Error unmarshalling liked tracks")
		return LikedTracksResultItem{}, err
	}

	return likedTracks.Items[0], nil
}

func (t SpotifyToken) GetEmbed(c context.Context, oauth2Conf *oauth2.Config) string {
	track, err := t.getRandomLikedTrack(c, oauth2Conf)
	if err != nil {
		return ""
	}

	// Get Embed HTML
	embedHTML, err := track.Track.ExternalUrls.getEmbed()
	if err != nil {
		return ""
	}
	return embedHTML.Html
}
