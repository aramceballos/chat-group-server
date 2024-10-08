package channel

import (
	"github.com/aramceballos/chat-group-server/pkg/entities"
)

type Service interface {
	FetchChannels() ([]entities.Channel, error)
	FetchChannelById(id string) (entities.Channel, error)
	CreateChannel(channel entities.CreateChannelInput) error
	JoinChannel(userId string, input entities.JoinChannelInput) error
	LeaveChannel(userId string, input entities.JoinChannelInput) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo,
	}
}

func (s *service) FetchChannels() ([]entities.Channel, error) {
	channels, err := s.repo.FetchChannels()
	if err != nil {
		return nil, err
	}

	return channels, nil
}

func (s *service) FetchChannelById(id string) (entities.Channel, error) {
	channel, err := s.repo.FetchChannelById(id)
	if err != nil {
		return entities.Channel{}, err
	}

	return channel, nil
}

func (s *service) CreateChannel(channel entities.CreateChannelInput) error {
	return s.repo.CreateChannel(channel)
}

func (s *service) JoinChannel(userId string, input entities.JoinChannelInput) error {
	return s.repo.JoinChannel(userId, input)
}

func (s *service) LeaveChannel(userId string, input entities.JoinChannelInput) error {
	return s.repo.LeaveChannel(userId, input)
}
