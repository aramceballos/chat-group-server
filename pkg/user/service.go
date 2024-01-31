package user

import (
	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Service interface {
	FetchUsers() ([]entities.User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo,
	}
}

func (s *service) FetchUsers() ([]entities.User, error) {
	users, err := s.repo.FetchUsers()
	if err != nil {
		return nil, err
	}

	return users, nil
}
