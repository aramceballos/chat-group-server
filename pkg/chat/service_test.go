package chat

import (
	"errors"
	"testing"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/stretchr/testify/assert"
)

type mockRepository struct {
	membershipResult bool
	membershipError  error
	user             entities.User
	userError        error
	message          entities.Message
	messageError     error
}

func (mr mockRepository) CheckUserMembership(channelId int, userId int) (bool, error) {
	return mr.membershipResult, mr.membershipError
}

func (mr mockRepository) FetchUserById(userId int) (entities.User, error) {
	return mr.user, mr.userError
}

func (mr mockRepository) InsertMessage(channelId int, userId int, msgBody []byte) (entities.Message, error) {
	return mr.message, mr.messageError
}

func (mr mockRepository) Close() error {
	return nil
}

func TestNewService(t *testing.T) {
	mockRepo := mockRepository{}
	s := NewService(mockRepo)
	assert.NotNil(t, s)
}

func TestCheckUserMembership(t *testing.T) {
	t.Run("membership exists", func(t *testing.T) {
		mockRepo := mockRepository{membershipResult: true}
		s := NewService(mockRepo)
		exists, err := s.CheckUserMembership(1, 1)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("membership does not exist", func(t *testing.T) {
		mockRepo := mockRepository{membershipResult: false}
		s := NewService(mockRepo)
		exists, err := s.CheckUserMembership(1, 1)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("membership returns error", func(t *testing.T) {
		mockRepo := mockRepository{membershipError: errors.New("db error")}
		s := NewService(mockRepo)
		exists, err := s.CheckUserMembership(1, 1)
		assert.Error(t, err)
		assert.False(t, exists)
	})
}

func TestFetchUserByIdService(t *testing.T) {
	t.Run("user found", func(t *testing.T) {
		user := entities.User{ID: 1, Name: "John"}
		mockRepo := mockRepository{user: user}
		s := NewService(mockRepo)
		result, err := s.FetchUserById(1)
		assert.NoError(t, err)
		assert.Equal(t, user, result)
	})

	t.Run("user error", func(t *testing.T) {
		mockRepo := mockRepository{userError: errors.New("user not found")}
		s := NewService(mockRepo)
		result, err := s.FetchUserById(1)
		assert.Error(t, err)
		assert.Equal(t, entities.User{}, result)
	})
}

func TestInsertMessageService(t *testing.T) {
	t.Run("message inserted", func(t *testing.T) {
		msg := entities.Message{ID: 1, ChannelID: 1, UserID: 1}
		mockRepo := mockRepository{message: msg}
		s := NewService(mockRepo)
		result, err := s.InsertMessage(1, 1, []byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, msg, result)
	})

	t.Run("insert error", func(t *testing.T) {
		mockRepo := mockRepository{messageError: errors.New("insert failed")}
		s := NewService(mockRepo)
		result, err := s.InsertMessage(1, 1, []byte("hello"))
		assert.Error(t, err)
		assert.Equal(t, entities.Message{}, result)
	})
}
