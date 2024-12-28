# auth-code-flow-spotify

Golang sample application with Spotify API

## Features

- User authentication
- Spotify API authorization
- Show random liked song of the spotify user

## Technical Stack

- net/http
- gorm (w/ sqlite)
- golang.org/x/oauth2

## Deployment

1. Configure Spotify developer account and register your OAuth application.
1. Add `.env` file
   ```
   SPOTIFY_CLIENT_ID=<your-id>
   SPOTIFY_CLIENT_SECRET=<your-secret>
   REDIRECT_URL=http://localhost:8080/callback
   ```
1. Connect `localhost:8080`
