package handler

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/golang-jwt/jwt"
)

type TokenCreator struct {
	PrivateKey *rsa.PrivateKey
	Expiry     time.Duration
}

func (tc *TokenCreator) CreateAccessToken(usr *user.User) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		Subject:   usr.ID(),
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(tc.expiry()).Unix(),
	})

	return token.SignedString(tc.PrivateKey)
}

func (tc *TokenCreator) expiry() time.Duration {
	if tc.Expiry <= 0 {
		return 1 * time.Hour
	}

	return tc.Expiry

}

type TokenVerifier struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Expiry     time.Duration
}

func (tv *TokenVerifier) VerifyIdentify(tokenString string) (string, error) {
	var claim jwt.StandardClaims
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (interface{}, error) {
		return tv.PublicKey, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claim.Subject, nil
}
