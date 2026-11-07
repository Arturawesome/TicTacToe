package service

import (
	"context"
	"errors"
	"sync"
	"ticTacToe/internal/data"
	"ticTacToe/internal/game"
	"ticTacToe/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNoActiveGame    = errors.New("no active game")
	ErrNotHumanTurn    = errors.New("current player is not human")
	ErrNotYourTurn     = errors.New("not your turn")
	ErrNotComputerTurn = errors.New("current player is not computer")
	ErrInvalidOpponent = errors.New("invalid opponent type")
)

type GameStatus int

const (
	WaitingPlayers GameStatus = iota
	PlayerTurn
	Draw
	PlayerWon
)

type PlayerType int

const (
	HumanPlayer PlayerType = iota
	ComputerPlayer
)

type Player struct {
	ID   string
	Type PlayerType
	Mark game.Mark
}

type GameService struct {
	joinMu sync.Mutex
	g      *game.Game

	gameID string
	repo   data.GameRepository

	player1 Player
	player2 Player

	gameStatus GameStatus
}

func (s *GameService) Game() (*game.Game, error) {
	if s.g == nil {
		return nil, ErrNoActiveGame
	}

	return s.g, nil
}

func NewGameService(repo data.GameRepository) *GameService {
	return &GameService{
		repo: repo,

		player1: Player{
			Type: HumanPlayer,
			Mark: game.X,
		},
		player2: Player{
			Type: ComputerPlayer,
			Mark: game.O,
		},
	}
}

func (s *GameService) GetLeaderboard(ctx context.Context,
	limit int) ([]model.LeaderboardModel, error) {
	return s.repo.GetLeaderboard(
		ctx,
		limit,
		int(Draw),
		int(PlayerWon),
	)
}

func (s *GameService) GetFinishedGames(ctx context.Context,
	userID string) ([]model.GameModel, error) {
	return s.repo.GetFinishedByUserID(
		ctx,
		userID,
		int(Draw),
		int(PlayerWon),
	)
}

func (s *GameService) CurrentPlayer() Player {
	if s.g.CurrentTurn() == s.player1.Mark {
		return s.player1
	}
	return s.player2
}

func (s *GameService) GetAvailableGames(ctx context.Context) ([]model.GameModel, error) {
	return s.repo.GetAvailable(
		ctx,
		int(WaitingPlayers),
		int(HumanPlayer),
	)
}

func (s *GameService) JoinGame(ctx context.Context, gameID string, userID string) error {
	s.joinMu.Lock()
	defer s.joinMu.Unlock()

	if err := s.LoadGame(ctx, gameID); err != nil {
		return err
	}
	if s.gameStatus != WaitingPlayers {
		return errors.New("game is not waiting for players")
	}

	if s.player2.Type != HumanPlayer {
		return errors.New("cannot join computer game")
	}

	if s.player2.ID != "" {
		return errors.New("game already has second player")
	}

	if s.player1.ID == userID {
		return errors.New("cannot join your own game")
	}

	s.player2.ID = userID
	s.player2.Type = HumanPlayer
	s.updateGameStatus()

	return s.repo.Update(
		ctx,
		s.toModel(),
	)
}
func (s *GameService) LoadGame(
	ctx context.Context,
	gameID string,
) error {
	model, err := s.repo.Get(ctx, gameID)
	if err != nil {
		return err
	}

	g, err := data.ToDomain(model)
	if err != nil {
		return err
	}

	s.g = g
	s.gameID = model.ID

	s.player1 = Player{
		ID:   model.Player1ID,
		Type: HumanPlayer,
		Mark: game.Mark(model.Player1Mark),
	}

	s.player2 = Player{
		Type: PlayerType(model.OpponentType),
		Mark: game.Mark(model.Player2Mark),
	}

	if model.Player2ID != nil {
		s.player2.ID = *model.Player2ID
	}

	s.gameStatus = GameStatus(model.GameStatus)

	return nil
}

func (s *GameService) GetGameByID(
	ctx context.Context,
	gameID string,
) (model.GameModel, error) {
	return s.repo.Get(ctx, gameID)
}

func (s *GameService) StartGame(
	ctx context.Context,
	userID string,
	opponentType PlayerType,
) (string, error) {
	if opponentType != HumanPlayer &&
		opponentType != ComputerPlayer {
		return "", ErrInvalidOpponent
	}

	g := game.NewGame()
	gameID := uuid.NewString()

	player1 := Player{
		ID:   userID,
		Type: HumanPlayer,
		Mark: game.X,
	}

	player2 := Player{
		Type: opponentType,
		Mark: game.O,
	}

	s.g = g
	s.gameID = gameID
	s.player1 = player1
	s.player2 = player2
	s.updateGameStatus()

	if err := s.repo.Create(ctx, s.toModel()); err != nil {
		return "", err
	}
	return gameID, nil
}

func (s *GameService) MakeHumanMove(
	ctx context.Context,
	userID string,
	row, col int,
) error {
	if s.g == nil {
		return ErrNoActiveGame
	}

	currentPlayer := s.CurrentPlayer()

	if currentPlayer.Type != HumanPlayer {
		return ErrNotHumanTurn
	}

	if currentPlayer.ID != userID {
		return ErrNotYourTurn
	}

	if err := s.g.MakeMove(row, col); err != nil {
		return err
	}

	s.updateGameStatus()

	return s.repo.Update(
		ctx,
		s.toModel(),
	)
}

func (s *GameService) MakeComputerMove(
	ctx context.Context,
) error {
	if s.g == nil {
		return ErrNoActiveGame
	}

	currentPlayer := s.CurrentPlayer()
	if currentPlayer.Type != ComputerPlayer {
		return ErrNotComputerTurn
	}

	row, col := game.BestMove(s.g, currentPlayer.Mark)
	if row == -1 || col == -1 {
		return nil
	}

	if err := s.g.MakeMove(row, col); err != nil {
		return err
	}

	s.updateGameStatus()

	return s.repo.Update(
		ctx,
		s.toModel(),
	)
}

func (s *GameService) updateGameStatus() {
	if s.g == nil {
		s.gameStatus = WaitingPlayers
		return
	}

	if s.player2.Type == HumanPlayer && s.player2.ID == "" {
		s.gameStatus = WaitingPlayers
		return
	}

	if s.g.Winner() != game.Empty {
		s.gameStatus = PlayerWon
		return
	}

	if s.g.IsFull() {
		s.gameStatus = Draw
		return
	}

	s.gameStatus = PlayerTurn
}

func (s *GameService) Status() GameStatus {
	return s.gameStatus
}

func (s *GameService) WinnerPlayer() Player {
	if s.g == nil {
		return Player{}
	}

	winnerMark := s.g.Winner()

	if winnerMark == s.player1.Mark {
		return s.player1
	}

	if winnerMark == s.player2.Mark {
		return s.player2
	}

	return Player{}
}

func (s *GameService) toModel() model.GameModel {
	winnerID := ""

	if s.gameStatus == PlayerWon {
		winner := s.WinnerPlayer()
		winnerID = winner.ID
	}

	return data.ToModel(
		s.gameID,
		s.g,
		int(s.gameStatus),
		winnerID,
		s.player1.ID,
		s.player1.Mark,
		s.player2.ID,
		s.player2.Mark,
		int(s.player2.Type),
	)
}
