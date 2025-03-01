package handler

import (
	"github.com/SawitProRecruitment/UserService/handler/app"
	"github.com/SawitProRecruitment/UserService/repository"
)

type Server struct {
	Repository  repository.RepositoryInterface
	AuthService *app.AuthService
	UserService *app.UserService
}

type NewServerOptions struct {
	Repository repository.RepositoryInterface
}

func NewServer(opts NewServerOptions) *Server {
	return &Server{
		Repository:  opts.Repository,
		AuthService: app.NewAuthService(opts.Repository),
		UserService: app.NewUserService(opts.Repository),
	}
}
