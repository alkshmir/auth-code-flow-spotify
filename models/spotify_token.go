package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type SpotifyToken struct {
	gorm.Model
	UserID       uint
	User         User
	ID           uint      `gorm:"primaryKey"`
	AccessToken  string    `gorm:"not null"`
	RefreshToken string    `gorm:"not null"`
	Expiry       time.Time `json:"expiry,omitempty"`
	ExpiresIn    int64     `json:"expires_in,omitempty"`
}

func (t SpotifyToken) GenerateOauthToken() oauth2.Token {
	return oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
		ExpiresIn:    t.ExpiresIn,
	}
}

func CreateSpotifyToken(token oauth2.Token, user User) SpotifyToken {
	return SpotifyToken{
		UserID:       user.ID,
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		ExpiresIn:    token.ExpiresIn,
	}
}

func (t *SpotifyToken) UpdateFromOauth2Token(token oauth2.Token) {
	t.AccessToken = token.AccessToken
	t.RefreshToken = token.RefreshToken
	t.Expiry = token.Expiry
	t.ExpiresIn = token.ExpiresIn
}

// Refresh access token
func (t *SpotifyToken) Refresh(c context.Context, oauth2Conf *oauth2.Config) error {
	token := t.GenerateOauthToken()
	newToken, err := oauth2Conf.TokenSource(c, &token).Token()
	if err != nil {
		fmt.Println("Error refreshing token")
		return err
	}
	if t.AccessToken == newToken.AccessToken {
		fmt.Println("AccessToken is not refreshed")
	}
	if t.RefreshToken == newToken.RefreshToken {
		fmt.Println("RefreshToken is not refreshed")
	}
	// Save the new token
	t.UpdateFromOauth2Token(*newToken)

	return nil
}

func (t SpotifyToken) getOauthClient(c context.Context, oauth2Conf *oauth2.Config) (*http.Client, error) {
	token := oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
	}
	return oauth2Conf.Client(c, &token), nil
	//client := oauth2.NewClient(c, oauth2Conf.TokenSource(c, &token))
	//return client, nil
}

func (t SpotifyToken) getRandomLikedTrack(c context.Context, oauth2Conf *oauth2.Config) (LikedTracksResultItem, error) {
	client, err := t.getOauthClient(c, oauth2Conf)
	if err != nil {
		fmt.Println("Error getting oauth2 client")
		return LikedTracksResultItem{}, err
	}

	// call spotify API
	resp, err := client.Get("https://api.spotify.com/v1/me/tracks?limit=1&offset=0")
	if err != nil {
		fmt.Println("Error calling spotify API")
		return LikedTracksResultItem{}, err
	}
	fmt.Println(resp.StatusCode)

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
