package app

import (
	"errors"

	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/SawitProRecruitment/UserService/repository"
)

var ErrPhoneNumberAlreadyTaken = errors.New("phone number already taken")

type UserService struct {
	userRepo repository.RepositoryInterface
}

func NewUserService(userRepo repository.RepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (us *UserService) RegisterUser(phoneNumber, fullName, password string) (*user.User, error) {
	usr, err := user.NewWithPassword(user.NextID(), phoneNumber, fullName, password)
	if err != nil {
		return nil, err
	}

	err = us.userRepo.Store(usr)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (us *UserService) GetProfile(id string) (*user.User, error) {
	usr, err := us.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return usr, nil
}

func (us *UserService) UpdateProfile(id string, fullName, phoneNumber *string) error {
	usr, err := us.userRepo.GetByID(id)
	if err != nil {
		return err
	}

	var modified bool
	if fullName != nil && *fullName != usr.FullName() {
		err = usr.ChangeFullName(*fullName)
		if err != nil {
			return err
		}

		modified = true
	}

	if phoneNumber != nil && *phoneNumber != usr.PhoneNumber() {
		other, err := us.userRepo.GetByPhoneNumber(*phoneNumber)
		if err != nil {
			return err
		}

		// Ensure the new phone number are not taken yet
		if other != nil && other.ID() != usr.ID() {
			return ErrPhoneNumberAlreadyTaken
		}

		err = usr.ChangePhoneNumber(*phoneNumber)
		if err != nil {
			return err
		}

		modified = true
	}

	if !modified {
		return nil
	}

	return us.userRepo.Update(usr)
}
