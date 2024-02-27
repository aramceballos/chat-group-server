package auth

import (
	"errors"
	"net/mail"
	"time"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func valid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type Service interface {
	Login(input entities.LoginInput) (string, error)
	Signup(input entities.SignupInput) (string, error)
	Me(userId string) (entities.User, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo,
	}
}

func (s *service) Login(input entities.LoginInput) (string, error) {
	var ud entities.UserData

	identity := input.Identity
	pass := input.Password
	user, email, err := new(entities.User), new(entities.User), *new(error)

	if valid(identity) {
		email, err = s.repo.GetUserByEmail(identity)
		if err != nil {
			return "", errors.New("user not found")
		}
		ud = entities.UserData{
			ID:       email.ID,
			UserName: email.UserName,
			Email:    email.Email,
			Password: email.Password,
		}
	} else {
		user, err = s.repo.GetUserByUsername(identity)
		if err != nil {
			return "", errors.New("user not found")
		}
		ud = entities.UserData{
			ID:       user.ID,
			UserName: user.UserName,
			Email:    user.Email,
			Password: user.Password,
		}
	}

	if email == nil && user == nil {
		return "", errors.New("user not found")
	}

	if !CheckPasswordHash(pass, ud.Password) {
		return "", errors.New("invalid password")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = ud.UserName
	claims["user_id"] = ud.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *service) Signup(input entities.SignupInput) (string, error) {
	user := entities.User{
		Name:      input.Name,
		UserName:  input.UserName,
		Email:     input.Email,
		Password:  input.Password,
		AvatarURL: input.AvatarURL,
	}

	if user.Name == "" || user.UserName == "" || user.Email == "" || user.Password == "" {
		return "", errors.New("all fields are required")
	}

	if !valid(user.Email) {
		return "", errors.New("invalid email")
	}

	_, err := s.repo.GetUserByEmail(user.Email)
	if err == nil {
		return "", errors.New("email already in use")
	}

	_, err = s.repo.GetUserByUsername(user.UserName)
	if err == nil {
		return "", errors.New("username already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user.Password = string(hash)

	err = s.repo.CreateUser(user)
	if err != nil {
		return "", err
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.UserName
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (s *service) Me(userId string) (entities.User, error) {
	user, err := s.repo.GetUserById(userId)
	if err != nil {
		return entities.User{}, err
	}

	return *user, nil
}
