package services

import (
	"net/http"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

// SessionService is responsible for anything user related when they're not logged in such as
// logging in, registering, resetting passwords, requesting new tokens, etc...
type SessionService interface {
	Login(ctx Context) error
	Register(ctx Context) error
	Refresh(ctx Context) error
}

// NewSessionService initializer
func NewSessionService(sessionInteractor usecases.SessionInteractor) SessionService {
	return &sessionService{
		SessionInteractor: sessionInteractor,
	}
}

type sessionService struct {
	SessionInteractor usecases.SessionInteractor
}

// SessionLoginBody is the data that's needed to log in
type SessionLoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *sessionService) Login(ctx Context) error {
	b := &SessionLoginBody{}
	err := ctx.Bind(b)
	if err != nil {
		return domain.WrapError(err)
	}

	user, token, err := s.SessionInteractor.CreateSession(b.Email, b.Password)
	if err != nil {
		ctx.NoContent(http.StatusUnauthorized)
		return domain.WrapError(err)
	}

	res := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	return ctx.JSON(http.StatusOK, res)
}

func (s *sessionService) Register(ctx Context) error {
	user := &domain.User{}
	err := ctx.Bind(user)
	if err != nil {
		return domain.WrapError(err)
	}

	user.Role = domain.RoleUser
	user.Preferences = &domain.Preferences{}

	err = s.SessionInteractor.CreateUser(*user)
	if err != nil {
		return domain.WrapError(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

func (s *sessionService) Refresh(ctx Context) error {
	sessionUser, err := ctx.User()
	if err != nil {
		return domain.WrapError(err)
	}

	user, token, err := s.SessionInteractor.RefreshSession(*sessionUser)
	if err != nil {
		ctx.NoContent(http.StatusUnauthorized)
		return domain.WrapError(err)
	}

	res := map[string]interface{}{
		"token": token,
		"user":  user,
	}

	return ctx.JSON(http.StatusOK, res)
}
