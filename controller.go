package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"text/template"

	"github.com/alkshmir/auth-code-flow-spotify/models"
	"github.com/alkshmir/auth-code-flow-spotify/oauth"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Session struct {
	UserID uint
}

var (
	sessionStore = make(map[string]Session)
	sessionMutex sync.Mutex
)

// Register a new session and return the session ID
func newSession(userID uint) string {
	sessionID := uuid.NewString()
	sessionMutex.Lock()
	sessionStore[sessionID] = Session{UserID: userID}
	sessionMutex.Unlock()
	return sessionID
}

func setCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	})
}

// Handler for user registration
func registerHandler(db *gorm.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`
					<!DOCTYPE html>
					<html lang="en">
					<head>
						<meta charset="UTF-8">
						<title>Register</title>
					</head>
					<body>
						<h1>Register</h1>
						<form action="/register" method="post">
							<input type="text" name="username" placeholder="Username" required>
							<input type="password" name="password" placeholder="Password" required>
							<button type="submit">Register</button>
						</form>
						<a href="/login">Login</a>
					</body>
					</html>
				`))
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		user := models.User{Username: username, Password: string(hashedPassword)}
		if err := db.Create(&user).Error; err != nil {
			http.Error(w, "Error saving user", http.StatusInternalServerError)
			return
		}

		sessionId := newSession(user.ID)
		setCookie(w, sessionId)

		http.Redirect(w, r, "/hello", http.StatusFound)
	}
}

func loginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`
				<!DOCTYPE html>
				<html lang="en">
				<head>
					<meta charset="UTF-8">
					<title>Login</title>
				</head>
				<body>
					<h1>Login</h1>
					<form action="/login" method="post">
						<input type="text" name="username" placeholder="Username" required>
						<input type="password" name="password" placeholder="Password" required>
						<button type="submit">Login</button>
					</form>
					<a href="/register">Register</a>
				</body>
				</html>
			`))
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		sessionID := newSession(user.ID)
		setCookie(w, sessionID)

		http.Redirect(w, r, "/hello", http.StatusFound)
	}
}

// Auth middleware
func requireAuth(db *gorm.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			// Unauthorized
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		sessionMutex.Lock()
		session, exists := sessionStore[cookie.Value]
		sessionMutex.Unlock()

		if !exists {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		var user models.User
		if err := db.First(&user, session.UserID).Error; err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		var spotifyToken models.SpotifyToken
		if err := db.Where("user_id = ?", user.ID).First(&spotifyToken).Error; err != nil {
			fmt.Println("Spotify token not found")
			ctx = context.WithValue(ctx, "SpotifyToken", nil)
		} else {
			err := spotifyToken.Refresh(ctx, oauth.Oauth2Config())
			if err != nil {
				fmt.Println("Error refreshing token")
				ctx = context.WithValue(ctx, "SpotifyToken", nil)
			} else {
				// save token to DB
				if err := db.Save(spotifyToken).Error; err != nil {
					fmt.Println("Error saving token")
					ctx = context.WithValue(ctx, "SpotifyToken", nil)
				} else {
					fmt.Println("Token refreshed")
					ctx = context.WithValue(ctx, "SpotifyToken", spotifyToken)
				}
			}
		}
		next(w, r.WithContext(ctx))
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// コンテキストからユーザ情報を取得
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	spotifyToken, hasSpotifyToken := r.Context().Value("SpotifyToken").(models.SpotifyToken)

	var songEmbed string
	if hasSpotifyToken {
		fmt.Println("User has Spotify token")
		songEmbed = spotifyToken.GetEmbed(r.Context(), oauth.Oauth2Config())
	}

	data := struct {
		Username        string
		HasSpotifyToken bool
		LuckySongLink   string
		ApiLink         string
	}{
		Username:        user.Username,
		HasSpotifyToken: hasSpotifyToken,
		LuckySongLink:   songEmbed,
		ApiLink:         "/connect-spotify",
	}

	fmt.Println(spotifyToken)

	tmpl := template.Must(template.New("hello").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Hello</title>
	</head>
	<body>
		<h1>Hello</h1>
		<p>Hello, {{.Username}}</p>
		{{if .HasSpotifyToken}}
			<p>Lucky Song: {{ .LuckySongLink }}</p>
		{{else}}
			<p>Connect your Spotify account: <a href="{{.ApiLink}}">API Link</a></p>
		{{end}}
	</body>
	</html>
`))

	// Show username
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	tmpl.Execute(w, data)
}
