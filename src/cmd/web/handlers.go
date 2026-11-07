package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"ticTacToe/internal/model"
	"ticTacToe/internal/service"
)

func (app *application) GameHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	games, err := app.gameService.GetFinishedGames(r.Context(), userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(games); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) GameLeaderboard(
	w http.ResponseWriter,
	r *http.Request,
) {
	nString := r.URL.Query().Get("n")

	n, err := strconv.Atoi(nString)
	if err != nil {
		http.Error(w, "invalid n", http.StatusBadRequest)
		return
	}

	if n <= 0 {
		http.Error(w, "n must be greater than zero", http.StatusBadRequest)
		return
	}

	players, err := app.gameService.GetLeaderboard(r.Context(), n)
	if err != nil {
		http.Error(w, "cannot get leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(players)
	if err != nil {
		http.Error(w, "cannot encode response", http.StatusInternalServerError)
		return
	}
}

func (app *application) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	refreshReques := model.RefreshJwtRequest{}

	body := r.Body
	err := json.NewDecoder(body).Decode(&refreshReques)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	responce, err := app.authService.RefreshAccessToken(refreshReques)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responce)
	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) RefreshRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshReqguest := model.RefreshJwtRequest{}
	body := r.Body
	err := json.NewDecoder(body).Decode(&refreshReqguest)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	responce, err := app.authService.RefreshRefreshToken(refreshReqguest)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responce)
	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := app.userService.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"id":    user.ID,
		"login": user.Login,
	})
}

func (app *application) GetGame(
	w http.ResponseWriter,
	r *http.Request,
) {
	gameID := r.PathValue("id")
	if gameID == "" {
		http.Error(w, "game id required", http.StatusBadRequest)
		return
	}

	model, err := app.gameService.GetGameByID(
		r.Context(),
		gameID,
	)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(model); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) JoinGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	gameID := r.PathValue("id")
	if gameID == "" {
		http.Error(w, "game id required", http.StatusBadRequest)
		return
	}

	if err := app.gameService.JoinGame(
		r.Context(),
		gameID,
		userID,
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) AvailableGames(w http.ResponseWriter, r *http.Request) {
	games, err := app.gameService.GetAvailableGames(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(games); err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	req := model.JwtRequest{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	}
	ok, err := app.authService.Registration(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !ok {
		http.Error(w, "registration failed", http.StatusBadRequest)
		return
	}
	credentials := req.Login + ":" + req.Password

	authBase64 := base64.StdEncoding.EncodeToString(
		[]byte(credentials),
	)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, authBase64)

	w.WriteHeader(http.StatusCreated)
}

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {

	jwtRequest := model.JwtRequest{}
	body := r.Body
	err := json.NewDecoder(body).Decode(&jwtRequest)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	jwtResponce, err := app.authService.Login(jwtRequest)
	if err != nil {
		http.Error(w, "INCORRECT LOGIN OR PASSWORD", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(jwtResponce)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (app *application) CreateGame(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	opponenType, err := strconv.Atoi(r.FormValue("opponent"))
	if err != nil {
		http.Error(w, "invalid opponent", http.StatusBadRequest)
		return
	}
	gameID, err := app.gameService.StartGame(r.Context(), userID, service.PlayerType(opponenType))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"id": gameID,
	})

}

func (app *application) Move(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := userIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	gameID := r.PathValue("id")
	if gameID == "" {
		http.Error(w, "game id required", http.StatusBadRequest)
		return
	}

	if err := app.gameService.LoadGame(
		r.Context(),
		gameID,
	); err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	status := app.gameService.Status()

	if status == service.PlayerWon ||
		status == service.Draw {
		http.Error(w, "game already finished", http.StatusBadRequest)
		return
	}

	if status == service.WaitingPlayers {
		http.Error(w, "waiting for second player", http.StatusBadRequest)
		return
	}

	currentPlayer := app.gameService.CurrentPlayer()

	if currentPlayer.Type == service.HumanPlayer {
		row, err := strconv.Atoi(r.FormValue("i"))
		if err != nil {
			http.Error(w, "invalid row", http.StatusBadRequest)
			return
		}

		col, err := strconv.Atoi(r.FormValue("j"))
		if err != nil {
			http.Error(w, "invalid column", http.StatusBadRequest)
			return
		}

		if err := app.gameService.MakeHumanMove(
			r.Context(),
			userID,
			row,
			col,
		); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if app.gameService.Status() == service.PlayerTurn {
		currentPlayer = app.gameService.CurrentPlayer()

		if currentPlayer.Type == service.ComputerPlayer {
			if err := app.gameService.MakeComputerMove(
				r.Context(),
			); err != nil {
				app.serverError(w, r, err)
				return
			}
		}
	}

	model, err := app.gameService.GetGameByID(
		r.Context(),
		gameID,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(model); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) GetUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.PathValue("id")

	if userID == "" {
		http.Error(
			w,
			"user id required",
			http.StatusBadRequest,
		)
		return
	}

	user, err := app.userService.GetByID(
		r.Context(),
		userID,
	)
	if err != nil {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	response := struct {
		ID    string `json:"id"`
		Login string `json:"login"`
	}{
		ID:    user.ID,
		Login: user.Login,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

}
