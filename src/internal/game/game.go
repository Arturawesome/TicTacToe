package game

import "errors"

var (
	ErrInvalidPosition = errors.New("invalid position")
	ErrGameFinished    = errors.New("game already finished")
	ErrCellOccupied    = errors.New("cell occupied")
)

type Mark int

const (
	Empty Mark = iota
	X
	O
)

type Board [3][3]Mark

type Game struct {
	board       Board
	currentTurn Mark
}

type Status int

const (
	InProgress Status = iota
	XWon
	OWon
	Draw
)

func NewGame() *Game {
	return &Game{
		currentTurn: X,
	}
}

func (g *Game) Reset() {
	g.currentTurn = X
	g.board = Board{}
}

func (g *Game) copy() *Game {
	return &Game{
		board:       g.board,
		currentTurn: g.currentTurn,
	}
}

func RestoreGame(board Board, currentTurn Mark) *Game {
	return &Game{
		board:       board,
		currentTurn: currentTurn,
	}
}
