package repository

import (
	"database/sql"
	"errors"

	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/lib/pq"
	"github.com/rs/xid"
)

const (
	errCodeUniqueViolation = pq.ErrorCode("23505")
)

var ErrUniqueViolation = errors.New("unique violation")

func (r *Repository) Store(u *user.User) error {
	_id, err := xid.FromString(u.ID())
	if err != nil {
		return err
	}

	pwdHash, pwdSalt := u.Password()
	_, err = r.Db.Exec("INSERT INTO users (id, phone_number, full_name, password_hash, password_salt) VALUES ($1, $2, $3, $4, $5)", _id, u.PhoneNumber(), u.FullName(), pwdHash, pwdSalt)

	// Detect unique constraint violation!
	var pgerr *pq.Error
	if errors.As(err, &pgerr) {
		if pgerr.Code == errCodeUniqueViolation {
			return ErrUniqueViolation
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetByID(id string) (*user.User, error) {
	var (
		phoneNumber string
		fullName    string
		pwdHash     []byte
		pwdSalt     []byte
	)

	_id, err := xid.FromString(id)
	if err != nil {
		return nil, err
	}

	err = r.Db.QueryRow("SELECT phone_number, full_name, password_hash, password_salt FROM users WHERE id = $1", _id).Scan(
		&phoneNumber,
		&fullName,
		&pwdHash,
		&pwdSalt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user.New(id, phoneNumber, fullName, pwdHash, pwdSalt)
}

func (r *Repository) GetByPhoneNumber(phoneNumber string) (*user.User, error) {
	var (
		_id      xid.ID
		fullName string
		pwdHash  []byte
		pwdSalt  []byte
	)
	err := r.Db.QueryRow("SELECT id, full_name, password_hash, password_salt FROM users WHERE phone_number = $1", phoneNumber).Scan(
		&_id,
		&fullName,
		&pwdHash,
		&pwdSalt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return user.New(_id.String(), phoneNumber, fullName, pwdHash, pwdSalt)
}

func (r *Repository) Update(u *user.User) error {
	_id, err := xid.FromString(u.ID())
	if err != nil {
		return err
	}

	pwdHash, pwdSalt := u.Password()
	res, err := r.Db.Exec("UPDATE users SET phone_number = $1, full_name = $2, password_hash = $3, password_salt = $4 WHERE id = $5",
		u.PhoneNumber(),
		u.FullName(),
		pwdHash,
		pwdSalt,
		_id)

	// Detect unique constraint violation!
	var pgerr *pq.Error
	if errors.As(err, &pgerr) {
		if pgerr.Code == errCodeUniqueViolation {
			return ErrUniqueViolation
		}
	}

	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}
