package data

// user.go
import (
	"context"
	"ticTacToe/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// type model.UserModel struct {
// 	ID             string `db:"id"`
// 	Login          string `db:"login"`
// 	HashedPassword string `db:"hashed_password"`
// }

type UserRepository interface {
	Create(ctx context.Context, user model.UserModel) error
	GetByLogin(ctx context.Context, login string) (model.UserModel, error)
	GetByID(ctx context.Context, id string) (model.UserModel, error)
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (repo *PostgresUserRepository) Create(
	ctx context.Context, user model.UserModel) error {
	query := `
    INSERT INTO users (
        id,
        login,
        hashed_password
    )
    VALUES ($1, $2, $3)
`

	_, err := repo.db.Exec(
		ctx,
		query,
		user.ID,
		user.Login,
		user.HashedPassword,
	)
	return err
}

func (repo *PostgresUserRepository) GetByLogin(
	ctx context.Context,
	login string,
) (model.UserModel, error) {
	query := `
		SELECT
			id,
			login,
			hashed_password
		FROM users
		WHERE login = $1
	`

	var user model.UserModel

	err := repo.db.QueryRow(ctx, query, login).Scan(
		&user.ID,
		&user.Login,
		&user.HashedPassword,
	)

	if err != nil {
		return model.UserModel{}, err
	}

	return user, nil
}

func (repo *PostgresUserRepository) GetByID(
	ctx context.Context,
	id string,
) (model.UserModel, error) {
	query := `
		SELECT
			id,
			login,
			hashed_password
		FROM users
		WHERE id = $1
	`

	var user model.UserModel

	err := repo.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Login,
		&user.HashedPassword,
	)

	if err != nil {
		return model.UserModel{}, err
	}

	return user, nil
}
