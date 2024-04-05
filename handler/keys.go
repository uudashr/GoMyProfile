package handler

import (
	"crypto/rsa"
	"embed"

	"github.com/golang-jwt/jwt"
)

//go:embed *.pem
var keys embed.FS

func rsaPrivateKey(fs embed.FS, filename string) (*rsa.PrivateKey, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(data)
}

func rsaPublicKey(fs embed.FS, filename string) (*rsa.PublicKey, error) {
	data, err := fs.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(data)
}

var (
	_privateKey *rsa.PrivateKey
	_publicKey  *rsa.PublicKey
)

func init() {
	var err error
	_privateKey, err = rsaPrivateKey(keys, "private.pem")
	if err != nil {
		panic(err)
	}

	_publicKey, err = rsaPublicKey(keys, "public.pem")
	if err != nil {
		panic(err)
	}
}
