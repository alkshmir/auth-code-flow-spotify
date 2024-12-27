package main

import (
	"net/http"

	"github.com/alkshmir/auth-code-flow-spotify/models"
	"github.com/alkshmir/auth-code-flow-spotify/oauth"
)

func main() {
	db := models.InitDB()
	oauth.Init()

	http.HandleFunc("/register", registerHandler(db))
	http.HandleFunc("/login", loginHandler(db))
	http.Handle("/hello", requireAuth(db, helloHandler))
	http.Handle("/connect-spotify", requireAuth(db, oauth.ConnectSpotifyHandler))
	http.Handle("/callback", requireAuth(db, oauth.Oauth2CallbackHandler(db)))

	http.ListenAndServe(":8080", nil)
}
