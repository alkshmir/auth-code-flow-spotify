package oauth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alkshmir/auth-code-flow-spotify/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"os"

	"golang.org/x/oauth2"
)

func Init() error {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file, using env variable")
		return err
	}
	return nil
}

func Oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
		RedirectURL: os.Getenv("REDIRECT_URL"), // http://localhost:8080/callback
		Scopes:      []string{"user-library-read"},
	}
}

func ConnectSpotifyHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	state := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:  "oauthstate",
		Value: state,
		// Secure: true, // TODO: production env
		HttpOnly: true,
		Path:     "/",
	})

	url := Oauth2Config().AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Oauth2CallbackHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		cookie, err := r.Cookie("oauthstate")
		if err != nil {
			http.Error(w, "State cookie not found", http.StatusBadRequest)
			return
		}
		oauth2State := cookie.Value

		if r.URL.Query().Get("state") != oauth2State {
			fmt.Println(r.URL.Query().Get("state"))
			http.Error(w, "State parameter doesn't match", http.StatusBadRequest)
			return
		}

		// Get code
		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Println("auth code is empty")
			http.Error(w, "Auth code is empty", http.StatusInternalServerError)
			return
		}
		fmt.Println("auth code: ", code)

		token, err := Oauth2Config().Exchange(context.Background(), code)

		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "Failed to exchange token and code", http.StatusInternalServerError)
			return
		}

		// TODO: store token to secure storage
		fmt.Println("Access token: ", token.AccessToken)
		fmt.Println("Refresh token: ", token.RefreshToken)
		spotifyToken := models.CreateSpotifyToken(*token, user)
		if err := db.Create(&spotifyToken).Error; err != nil {
			http.Error(w, "Error storing token", http.StatusInternalServerError)
			return
		}
		// w.Write([]byte(`You are now authenticated. Try <a href=\"/track\">random liked songs</a>`))
		http.Redirect(w, r, "/hello", http.StatusFound)
	}
}
