package data

import (
	"context"
	"fmt"

	"ticTacToe/internal/game"
	"ticTacToe/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepository interface {
	Create(ctx context.Context, game model.GameModel) error
	Update(ctx context.Context, game model.GameModel) error
	Get(ctx context.Context, id string) (model.GameModel, error)
	GetAvailable(ctx context.Context, gameStatus int, opponentType int) ([]model.GameModel, error)

	GetFinishedByUserID(ctx context.Context, userID string, drawStatus int,
		wonStatus int) ([]model.GameModel, error)

	GetLeaderboard(
		ctx context.Context,
		limit int,
		drawStatus int,
		wonStatus int,
	) ([]model.LeaderboardModel, error)
}

type PostgresGameRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGameRepository(
	db *pgxpool.Pool,
) *PostgresGameRepository {
	return &PostgresGameRepository{
		db: db,
	}
}

func (r *PostgresGameRepository) GetLeaderboard(
	ctx context.Context,
	limit int,
	drawStatus int,
	wonStatus int,
) ([]model.LeaderboardModel, error) {

	query := `
		WITH player_results AS (
			SELECT
				player1_id AS user_id,
				CASE
					WHEN game_status = $2 THEN 'draw'
					WHEN winner = player1_id THEN 'win'
					ELSE 'loss'
				END AS result
			FROM games
			WHERE game_status IN ($2, $3)

			UNION ALL

			SELECT
				player2_id AS user_id,
				CASE
					WHEN game_status = $2 THEN 'draw'
					WHEN winner = player2_id THEN 'win'
					ELSE 'loss'
				END AS result
			FROM games
			WHERE
				game_status IN ($2, $3)
				AND player2_id IS NOT NULL
		),

		stats AS (
			SELECT
				user_id,

				COUNT(*) FILTER (
					WHERE result = 'win'
				) AS wins,

				COUNT(*) FILTER (
					WHERE result = 'loss'
				) AS losses,

				COUNT(*) FILTER (
					WHERE result = 'draw'
				) AS draws

			FROM player_results
			GROUP BY user_id
		)

		SELECT
			s.user_id,
			u.login,

			CASE
				WHEN s.losses + s.draws = 0
				THEN s.wins::DOUBLE PRECISION

				ELSE
					s.wins::DOUBLE PRECISION /
					(s.losses + s.draws)
			END AS win_ratio

		FROM stats s

		JOIN users u
			ON u.id = s.user_id

		ORDER BY
			win_ratio DESC,
			s.wins DESC

		LIMIT $1
	`

	rows, err := r.db.Query(
		ctx,
		query,
		limit,
		drawStatus,
		wonStatus,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaderboard []model.LeaderboardModel

	for rows.Next() {
		var player model.LeaderboardModel

		err := rows.Scan(
			&player.UserID,
			&player.Login,
			&player.WinRatio,
		)
		if err != nil {
			return nil, err
		}

		leaderboard = append(
			leaderboard,
			player,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return leaderboard, nil
}

func (r *PostgresGameRepository) GetFinishedByUserID(
	ctx context.Context,
	userID string,
	drawStatus int,
	wonStatus int,
) ([]model.GameModel, error) {
	query := `
		SELECT
			id,
			board,
			current_turn,
			game_status,
			winner,
			player1_id,
			player1_mark,
			player2_id,
			player2_mark,
			opponent_type,
			created_at
		FROM games
		WHERE
			(player1_id = $1 OR player2_id = $1)
			AND game_status IN ($2, $3)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		userID,
		drawStatus,
		wonStatus,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := make([]model.GameModel, 0)

	for rows.Next() {
		var model_ model.GameModel

		err := rows.Scan(
			&model_.ID,
			&model_.Board,
			&model_.CurrentTurn,
			&model_.GameStatus,
			&model_.Winner,
			&model_.Player1ID,
			&model_.Player1Mark,
			&model_.Player2ID,
			&model_.Player2Mark,
			&model_.OpponentType,
			&model_.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		games = append(games, model_)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}
func (r *PostgresGameRepository) Create(
	ctx context.Context,
	game model.GameModel,
) error {

	query := `
		INSERT INTO games (
			id,
			board,
			current_turn,
			game_status,
			winner,
			player1_id,
			player1_mark,
			player2_id,
			player2_mark,
			opponent_type
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		game.ID,
		game.Board,
		game.CurrentTurn,
		game.GameStatus,
		game.Winner,
		game.Player1ID,
		game.Player1Mark,
		game.Player2ID,
		game.Player2Mark,
		game.OpponentType,
	)

	return err
}

func (r *PostgresGameRepository) Get(
	ctx context.Context,
	id string,
) (model.GameModel, error) {

	query := `
		SELECT
			id,
			board,
			current_turn,
			game_status,
			winner,
			player1_id,
			player1_mark,
			player2_id,
			player2_mark,
			opponent_type,
			created_at
		FROM games
		WHERE id = $1
	`

	var model_ model.GameModel

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&model_.ID,
		&model_.Board,
		&model_.CurrentTurn,
		&model_.GameStatus,
		&model_.Winner,
		&model_.Player1ID,
		&model_.Player1Mark,
		&model_.Player2ID,
		&model_.Player2Mark,
		&model_.OpponentType,
		&model_.CreatedAt,
	)

	if err != nil {
		return model.GameModel{}, err
	}

	return model_, nil
}

func (r *PostgresGameRepository) Update(
	ctx context.Context,
	game model.GameModel,
) error {

	query := `
		UPDATE games
		SET
			board = $2,
			current_turn = $3,
			game_status = $4,
			winner = $5,
			player1_id = $6,
			player1_mark = $7,
			player2_id = $8,
			player2_mark = $9,
			opponent_type = $10
		WHERE id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		game.ID,
		game.Board,
		game.CurrentTurn,
		game.GameStatus,
		game.Winner,
		game.Player1ID,
		game.Player1Mark,
		game.Player2ID,
		game.Player2Mark,
		game.OpponentType,
	)

	return err
}

func ToModel(
	id string,
	g *game.Game,
	gameStatus int,
	winnerID string,
	player1ID string,
	player1Mark game.Mark,
	player2ID string,
	player2Mark game.Mark,
	opponentType int,
) model.GameModel {

	board := g.Board()

	flatBoard := make([]int, 0, 9)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			flatBoard = append(
				flatBoard,
				int(board[i][j]),
			)
		}
	}

	var winnerIDPtr *string

	if winnerID != "" {
		winnerIDPtr = &winnerID
	}

	var player2IDPtr *string

	if player2ID != "" {
		player2IDPtr = &player2ID
	}

	return model.GameModel{
		ID:          id,
		Board:       flatBoard,
		CurrentTurn: int(g.CurrentTurn()),
		GameStatus:  gameStatus,

		Winner: winnerIDPtr,

		Player1ID:   player1ID,
		Player1Mark: int(player1Mark),

		Player2ID:   player2IDPtr,
		Player2Mark: int(player2Mark),

		OpponentType: opponentType,
	}
}
func (r *PostgresGameRepository) GetAvailable(
	ctx context.Context,
	gameStatus int,
	opponentType int,
) ([]model.GameModel, error) {

	query := `
		SELECT
			id,
			board,
			current_turn,
			game_status,
			winner,
			player1_id,
			player1_mark,
			player2_id,
			player2_mark,
			opponent_type,
			created_at
		FROM games
		WHERE game_status = $1
		  AND opponent_type = $2
		  AND player2_id IS NULL
	`

	rows, err := r.db.Query(
		ctx,
		query,
		gameStatus,
		opponentType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []model.GameModel

	for rows.Next() {
		var model_ model.GameModel

		err := rows.Scan(
			&model_.ID,
			&model_.Board,
			&model_.CurrentTurn,
			&model_.GameStatus,
			&model_.Winner,
			&model_.Player1ID,
			&model_.Player1Mark,
			&model_.Player2ID,
			&model_.Player2Mark,
			&model_.OpponentType,
			&model_.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		games = append(games, model_)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

func ToDomain(
	model_ model.GameModel,
) (*game.Game, error) {
	if len(model_.Board) != 9 {
		return nil, fmt.Errorf(
			"invalid board size: %d",
			len(model_.Board),
		)
	}

	var board game.Board

	k := 0

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			board[i][j] = game.Mark(
				model_.Board[k],
			)
			k++
		}
	}

	currentTurn := game.Mark(
		model_.CurrentTurn,
	)

	return game.RestoreGame(
		board,
		currentTurn,
	), nil
}
