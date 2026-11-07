package game

func (g *Game) Board() Board {
	return g.board
}

func (g *Game) CurrentTurn() Mark {
	return g.currentTurn
}

func (g *Game) MakeMove(i int, j int) error {
	if i < 0 || i >= 3 || j < 0 || j >= 3 {
		return ErrInvalidPosition
	}
	if g.IsFinished() {
		return ErrGameFinished
	}
	if g.board[i][j] != Empty {
		return ErrCellOccupied
	}
	g.board[i][j] = g.currentTurn
	g.changePlayer()
	return nil
}

func (g *Game) changePlayer() {
	if g.currentTurn == X {
		g.currentTurn = O
		return
	}
	g.currentTurn = X
}

func (g *Game) Status() Status {
	winner := g.Winner()
	switch winner {
	case X:
		return XWon
	case O:
		return OWon
	}
	if g.IsFull() {
		return Draw
	}
	return InProgress

}
func (g *Game) IsFull() bool {
	for row := range g.board {
		for col := range g.board[row] {
			if g.board[row][col] == Empty {
				return false
			}
		}
	}
	return true
}

func (g *Game) IsFinished() bool {
	return g.Status() != InProgress
}



func (g *Game) Winner() Mark {
	w := g.checkRows()
	if w != Empty {
		return w
	}
	w = g.checkCols()
	if w != Empty {
		return w
	}
	w = g.checkDiagonals()
	if w != Empty {
		return w
	}
	return Empty
}

func (g *Game) checkDiagonals() Mark {
	oFlag := true
	xFlag := true

	for i := 0; i < 3; i++ {
		oFlag = oFlag && (g.board[i][i] == O)
		xFlag = xFlag && (g.board[i][i] == X)
	}

	if oFlag == true {
		return O
	}
	if xFlag == true {
		return X
	}

	oFlag = true
	xFlag = true

	for i := 0; i < 3; i++ {
		oFlag = oFlag && (g.board[i][2-i] == O)
		xFlag = xFlag && (g.board[i][2-i] == X)
	}

	if oFlag == true {
		return O
	}
	if xFlag == true {
		return X
	}

	return Empty
}

func (g *Game) checkCols() Mark {
	for j := 0; j < 3; j++ {
		oFlag := true
		xFlag := true

		for i := 0; i < 3; i++ {
			oFlag = oFlag && (g.board[i][j] == O)
			xFlag = xFlag && (g.board[i][j] == X)
		}
		if oFlag {
			return O
		}
		if xFlag {
			return X
		}
	}
	return Empty
}

func (g *Game) checkRows() Mark {
	for _, row := range g.board {
		oFlag := true
		xFlag := true

		for _, cell := range row {
			oFlag = oFlag && (cell == O)
			xFlag = xFlag && (cell == X)
		}

		if oFlag {
			return O
		}
		if xFlag {
			return X
		}
	}
	return Empty
}

