package user

import (
	"errors"
	"strings"

	"github.com/rs/xid"
)

type User struct {
	id           string
	phoneNumber  string
	fullName     string
	passwordHash []byte
	passwordSalt []byte
}

func New(id, phoneNumber, fullName string, passwordHash []byte, passwordSalt []byte) (*User, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}

	if !ValidPhoneNumberLength(phoneNumber) {
		return nil, errors.New("invalid phone number length")
	}

	if !ValidPhoneNumberPrefix(phoneNumber) {
		return nil, errors.New("invalid phone number prefix")
	}

	if !ValidFullNameLength(fullName) {
		return nil, errors.New("invalid full name length")
	}

	if len(passwordHash) == 0 {
		return nil, errors.New("empty password hash")
	}

	if len(passwordSalt) == 0 {
		return nil, errors.New("empty password salt")
	}

	return &User{
		id:           id,
		phoneNumber:  phoneNumber,
		fullName:     fullName,
		passwordHash: passwordHash,
		passwordSalt: passwordSalt,
	}, nil
}

func NewWithPassword(id, phoneNumber, fullName, password string) (*User, error) {
	if !ValidPasswordStrength(password) {
		return nil, errors.New("invalid password")
	}

	salt, err := genSalt(16)
	if err != nil {
		return nil, err
	}

	hash := hashPassword(password, salt)

	return New(id, phoneNumber, fullName, hash, salt)
}

func (u *User) ID() string {
	return u.id
}

func (u *User) PhoneNumber() string {
	return u.phoneNumber
}

func (u *User) ChangePhoneNumber(phoneNumber string) error {
	if !ValidPhoneNumberLength(phoneNumber) {
		return errors.New("invalid phone number length")
	}

	if !ValidPhoneNumberPrefix(phoneNumber) {
		return errors.New("invalid phone number prefix")
	}

	u.phoneNumber = phoneNumber
	return nil
}

func (u *User) FullName() string {
	return u.fullName
}

func (u *User) ChangeFullName(fullName string) error {
	if !ValidFullNameLength(fullName) {
		return errors.New("invalid full name length")
	}

	u.fullName = fullName
	return nil
}

func (u *User) Password() (hash, salt []byte) {
	return u.passwordHash, u.passwordSalt
}

func (u *User) VerifyPassword(plain string) bool {
	return verifyPassword(plain, u.passwordSalt, u.passwordHash)
}

func ValidPhoneNumberLength(phoneNumber string) bool {
	return len(phoneNumber) >= 10 && len(phoneNumber) <= 13
}

func ValidPhoneNumberPrefix(phoneNumber string) bool {
	return strings.HasPrefix(phoneNumber, "+62")
}

func ValidFullNameLength(fullName string) bool {
	return len(fullName) >= 3 && len(fullName) <= 60
}

func ValidPasswordStrength(password string) bool {
	// password must be at min 6 and max 64 characters
	if len(password) < 6 || len(password) > 64 {
		return false
	}

	// password contains at least one uppercase letter
	if !containsUppercase(password) {
		return false
	}

	// password contains at least one number
	if !containsNumber(password) {
		return false
	}

	// password contains at least one special character (non-alphanumeric)
	if !containsSpecialCharacter(password) {
		return false
	}

	return true
}

func NextID() string {
	return xid.New().String()
}

func containsUppercase(s string) bool {
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			return true
		}
	}
	return false
}

func containsNumber(s string) bool {
	for _, c := range s {
		if '0' <= c && c <= '9' {
			return true
		}
	}
	return false
}

func containsSpecialCharacter(s string) bool {
	for _, c := range s {
		if !('0' <= c && c <= '9') && !('A' <= c && c <= 'Z') && !('a' <= c && c <= 'z') {
			return true
		}
	}
	return false
}
