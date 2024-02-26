package user

import (
	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Service interface {
	FetchUsers() ([]entities.User, error)
	FetchUserById(id string) (entities.User, error)
	UpdateUser(userId string, user entities.UpdateUserInput) error
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

func (s *service) FetchUserById(id string) (entities.User, error) {
	user, err := s.repo.FetchUserById(id)
	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (s *service) UpdateUser(userId string, user entities.UpdateUserInput) error {
	err := s.repo.UpdateUser(userId, user)
	if err != nil {
		return err
	}

	return nil
}
