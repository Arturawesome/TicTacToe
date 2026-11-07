package model

type LeaderboardModel struct {
	UserID   string  `db:"user_id" json:"user_id"`
	Login    string  `db:"login" json:"login"`
	WinRatio float64 `db:"win_ratio" json:"win_ratio"`
}
