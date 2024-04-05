// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"github.com/SawitProRecruitment/UserService/handler/model/user"
)

type RepositoryInterface interface {
	Store(*user.User) error
	GetByID(id string) (*user.User, error)
	GetByPhoneNumber(phoneNumber string) (*user.User, error)
	Update(*user.User) error
}
