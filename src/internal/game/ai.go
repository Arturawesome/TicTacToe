package game

func minimax(g *Game, aiMark Mark) int {
	if g.IsFinished() {
		winner := g.Winner()
		if winner == aiMark {
			return 1
		}
		if winner != Empty {
			return -1
		}
		return 0
	}

	if g.currentTurn == aiMark {
		bestScore := -2
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if g.board[i][j] != Empty {
					continue
				}

				newState := g.copy()
				_ = newState.MakeMove(i, j)
				score := minimax(newState, aiMark)
				if score > bestScore {
					bestScore = score
				}

			}
		}
		return bestScore
	}

	bestScore := 2
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.board[i][j] != Empty {
				continue
			}

			newState := g.copy()
			_ = newState.MakeMove(i, j)
			score := minimax(newState, aiMark)
			if score < bestScore {
				bestScore = score
			}
		}
	}
	return bestScore

}

func BestMove(g *Game, aiMark Mark) (int, int) {
	if g.IsFinished() || g.currentTurn != aiMark {
		return -1, -1
	}

	bestScore := -2
	bestRow := -1
	bestCol := -1

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if g.board[i][j] != Empty {
				continue
			}

			newState := g.copy()
			_ = newState.MakeMove(i, j)
			score := minimax(newState, aiMark)
			if score > bestScore {
				bestScore = score
				bestRow = i
				bestCol = j
			}

		}
	}
	return bestRow, bestCol
}
