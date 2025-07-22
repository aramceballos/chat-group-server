package user

import (
	"errors"
	"testing"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	users       map[string]entities.User
	shouldError bool
	errorMsg    string
}

func NewMockRepository(users map[string]entities.User) *MockRepository {
	return &MockRepository{
		users:       users,
		shouldError: false,
	}
}

func NewMockRepositoryWithError(errorMsg string) *MockRepository {
	return &MockRepository{
		users:       make(map[string]entities.User),
		shouldError: true,
		errorMsg:    errorMsg,
	}
}

func (repo *MockRepository) Close() error {
	return nil
}

func (repo *MockRepository) FetchUsers() ([]entities.User, error) {
	if repo.shouldError {
		return nil, errors.New(repo.errorMsg)
	}

	var result []entities.User
	for _, user := range repo.users {
		result = append(result, user)
	}
	return result, nil
}

func (repo *MockRepository) FetchUserById(id string) (entities.User, error) {
	if repo.shouldError {
		return entities.User{}, errors.New(repo.errorMsg)
	}

	user, ok := repo.users[id]
	if !ok {
		return entities.User{}, errors.New("User not found")
	}

	return user, nil
}

func (repo *MockRepository) UpdateUser(userId string, user entities.UpdateUserInput) error {
	if repo.shouldError {
		return errors.New(repo.errorMsg)
	}

	existingUser, ok := repo.users[userId]
	if !ok {
		return errors.New("User not found")
	}

	// Update the user fields
	if user.Name != "" {
		existingUser.Name = user.Name
	}
	if user.Username != "" {
		existingUser.UserName = user.Username
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}

	repo.users[userId] = existingUser
	return nil
}

func TestNewService(t *testing.T) {
	mockRepo := NewMockRepository(map[string]entities.User{})
	service := NewService(mockRepo)
	assert.NotNil(t, service)
}

func TestService_FetchUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{
			"1": {
				ID:        1,
				Name:      "John Doe",
				UserName:  "johndoe",
				Email:     "john@example.com",
				Password:  "hashedPassword123",
				AvatarURL: "https://example.com/avatar1.jpg",
				Role:      "user",
				CreatedAt: "2023-01-01T00:00:00Z",
			},
			"2": {
				ID:        2,
				Name:      "Jane Smith",
				UserName:  "janesmith",
				Email:     "jane@example.com",
				Password:  "hashedPassword456",
				AvatarURL: "https://example.com/avatar2.jpg",
				Role:      "user",
				CreatedAt: "2023-01-02T00:00:00Z",
			},
		})

		service := NewService(mockRepo)
		users, err := service.FetchUsers()

		require.NoError(t, err)
		require.Len(t, users, 2)

		// Verify specific user details
		assert.Equal(t, "John Doe", users[0].Name)
		assert.Equal(t, "Jane Smith", users[1].Name)
		assert.Equal(t, int64(1), users[0].ID)
		assert.Equal(t, int64(2), users[1].ID)
	})

	t.Run("empty repository", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{})
		service := NewService(mockRepo)

		users, err := service.FetchUsers()

		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepositoryWithError("database connection failed")
		service := NewService(mockRepo)

		users, err := service.FetchUsers()

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "database connection failed")
	})
}

func TestService_FetchUserById(t *testing.T) {
	testUser := entities.User{
		ID:        1,
		Name:      "John Doe",
		UserName:  "johndoe",
		Email:     "john@example.com",
		Password:  "hashedPassword123",
		AvatarURL: "https://example.com/avatar.jpg",
		Role:      "user",
		CreatedAt: "2023-01-01T00:00:00Z",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{
			"1": testUser,
		})
		service := NewService(mockRepo)

		user, err := service.FetchUserById("1")

		require.NoError(t, err)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Name, user.Name)
		assert.Equal(t, testUser.UserName, user.UserName)
		assert.Equal(t, testUser.Email, user.Email)
		assert.Equal(t, testUser.AvatarURL, user.AvatarURL)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{})
		service := NewService(mockRepo)

		user, err := service.FetchUserById("999")

		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepositoryWithError("database timeout")
		service := NewService(mockRepo)

		user, err := service.FetchUserById("1")

		assert.Error(t, err)
		assert.Equal(t, entities.User{}, user)
		assert.Contains(t, err.Error(), "database timeout")
	})
}

func TestService_UpdateUser(t *testing.T) {
	initialUser := entities.User{
		ID:        1,
		Name:      "John Doe",
		UserName:  "johndoe",
		Email:     "john@example.com",
		Password:  "hashedPassword123",
		AvatarURL: "https://example.com/avatar.jpg",
		Role:      "user",
		CreatedAt: "2023-01-01T00:00:00Z",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{
			"1": initialUser,
		})
		service := NewService(mockRepo)

		updateInput := entities.UpdateUserInput{
			Name:     "John Updated",
			Username: "johnupdated",
			Email:    "john.updated@example.com",
		}

		err := service.UpdateUser("1", updateInput)
		require.NoError(t, err)

		// Verify the user was updated
		updatedUser, err := service.FetchUserById("1")
		require.NoError(t, err)
		assert.Equal(t, "John Updated", updatedUser.Name)
		assert.Equal(t, "johnupdated", updatedUser.UserName)
		assert.Equal(t, "john.updated@example.com", updatedUser.Email)
		assert.Equal(t, initialUser.ID, updatedUser.ID) // ID should remain unchanged
	})

	t.Run("partial update", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{
			"1": initialUser,
		})
		service := NewService(mockRepo)

		updateInput := entities.UpdateUserInput{
			Name: "Only Name Updated",
			// Username and Email not provided
		}

		err := service.UpdateUser("1", updateInput)
		require.NoError(t, err)

		// Verify only name was updated
		updatedUser, err := service.FetchUserById("1")
		require.NoError(t, err)
		assert.Equal(t, "Only Name Updated", updatedUser.Name)
		assert.Equal(t, initialUser.UserName, updatedUser.UserName) // Should remain unchanged
		assert.Equal(t, initialUser.Email, updatedUser.Email)       // Should remain unchanged
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo := NewMockRepository(map[string]entities.User{})
		service := NewService(mockRepo)

		updateInput := entities.UpdateUserInput{
			Name:     "New Name",
			Username: "newusername",
			Email:    "new@example.com",
		}

		err := service.UpdateUser("999", updateInput)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepositoryWithError("database write failed")
		service := NewService(mockRepo)

		updateInput := entities.UpdateUserInput{
			Name:     "New Name",
			Username: "newusername",
			Email:    "new@example.com",
		}

		err := service.UpdateUser("1", updateInput)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database write failed")
	})
}
