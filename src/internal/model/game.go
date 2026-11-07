package model

import "time"

type GameModel struct {
	ID          string `db:"id"`
	Board       []int  `db:"board"`
	CurrentTurn int    `db:"current_turn"`
	GameStatus  int    `db:"game_status"`

	Winner *string `db:"winner"`

	Player1ID   string `db:"player1_id"`
	Player1Mark int    `db:"player1_mark"`

	Player2ID   *string `db:"player2_id"`
	Player2Mark int     `db:"player2_mark"`

	OpponentType int       `db:"opponent_type"`
	CreatedAt    time.Time `db:"created_at"`
}
