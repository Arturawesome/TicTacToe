package model

type UserModel struct {
	ID             string `db:"id"`
	Login          string `db:"login"`
	HashedPassword string `db:"hashed_password"`
}
