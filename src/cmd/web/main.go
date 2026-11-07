package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"

	"ticTacToe/internal/data"
	"ticTacToe/internal/service"
)

type application struct {
	logger      *slog.Logger
	gameService *service.GameService
	authService *service.AuthService
	userService *service.UserService
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()
	dbURL := "postgres://ttt:ttt@localhost:5432/ttt"
	db, err := data.NewDatabase(ctx, dbURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	gameRepo := data.NewPostgresGameRepository(db.Pool)
	userRepo := data.NewPostgresUserRepository(db.Pool)

	gameService := service.NewGameService(gameRepo)
	userService := service.NewUserService(userRepo)
	provider := service.NewJwtProvider("access_key", "secret_key")
	authService := service.NewAuthService(provider, userService)

	app := &application{
		logger:      logger,
		gameService: gameService,
		userService: userService,
		authService: authService,
	}
	//
	mux := app.routes()
	logger.Info("starting server", "addr", *addr)
	logger.Info("SERVER IS STARTING...")

	err = http.ListenAndServe(*addr, mux)
	logger.Error("SERVER STOPPED", "error", err)

	logger.Error(err.Error())
	os.Exit(1)
}
