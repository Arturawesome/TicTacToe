package service

import (
	"context"
	"ticTacToe/internal/data"
	"ticTacToe/internal/model"
)

type UserService struct {
	repo data.UserRepository
}

func NewUserService(repo data.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context,
	user model.UserModel) error {
	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByLogin(ctx context.Context,
	login string) (model.UserModel, error) {
	return s.repo.GetByLogin(ctx, login)
}

func (s *UserService) GetByID(ctx context.Context,
	id string) (model.UserModel, error) {
	return s.repo.GetByID(ctx, id)
}
