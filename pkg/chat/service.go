package chat

import (
	"fmt"
	"log"

	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Service interface {
	CheckUserMembership(channelId int, userId int) (bool, error)
	FetchUserById(userId int) (entities.User, error)
	InsertMessage(channelId int, userId int, msgBody []byte) (entities.Message, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo,
	}
}

func (s *service) CheckUserMembership(channelId int, userId int) (bool, error) {
	exists, err := s.repo.CheckUserMembership(channelId, userId)
	if err != nil {
		log.Printf("[chat service error] error checking user membership: %s", err.Error())
		return false, fmt.Errorf("error checking user membership")
	}
	return exists, nil
}

func (s *service) FetchUserById(userId int) (entities.User, error) {
	user, err := s.repo.FetchUserById(userId)
	if err != nil {
		log.Printf("[chat service error] error fetching user by id: %s", err.Error())
		return entities.User{}, fmt.Errorf("error fetching user by")
	}
	return user, nil
}

func (s *service) InsertMessage(channelId int, userId int, msgBody []byte) (entities.Message, error) {
	message, err := s.repo.InsertMessage(channelId, userId, msgBody)
	if err != nil {
		log.Printf("[chat service error] error inserting message: %s", err.Error())
		return entities.Message{}, fmt.Errorf("error inserting message")
	}
	return message, nil
}
