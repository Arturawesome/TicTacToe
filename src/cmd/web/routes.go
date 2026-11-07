package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	//authenticator := NewUserAuthenticator(app.authService)

	mux.HandleFunc("POST /user/register", app.Register)
	mux.HandleFunc("POST /user/login", app.LoginHandler)
	mux.HandleFunc("POST /user/refresh/accessToken", app.RefreshAccessToken)
	mux.HandleFunc("POST /user/refresh/refreshToken", app.RefreshRefreshToken)

	handlerF := http.HandlerFunc(app.GetUser)
	mux.Handle("GET /user/{id}", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.GameLeaderboard)

	mux.Handle("GET /games/leaderboard", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.GameHistory)
	mux.Handle("GET /games/history", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.MeHandler)
	mux.Handle("GET /user/me", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.CreateGame)
	mux.Handle("POST /games", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.GetGame)
	mux.Handle("GET /games/{id}", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.JoinGame)
	mux.Handle("POST /games/{id}/join", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.Move)
	mux.Handle("POST /games/{id}/move", app.authJwtMiddleware(handlerF))

	handlerF = http.HandlerFunc(app.AvailableGames)
	mux.Handle("GET /games/available", app.authJwtMiddleware(handlerF))

	return mux
}
