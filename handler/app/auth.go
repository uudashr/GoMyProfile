package app

import (
	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/SawitProRecruitment/UserService/repository"
)

type AuthService struct {
	userRepo repository.RepositoryInterface
}

func NewAuthService(userRepo repository.RepositoryInterface) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (as *AuthService) Authenticate(phoneNumber, password string) (*user.User, error) {
	usr, err := as.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return nil, err
	}

	if usr == nil {
		return nil, AuthenticationError("user not found")
	}

	if !usr.VerifyPassword(password) {
		return nil, AuthenticationError("invalid password")
	}

	return usr, nil
}

type AuthenticationError string

func (ae AuthenticationError) Error() string {
	return string(ae)
}

type Claims struct {
	UserID string `json:"userId"`
}
