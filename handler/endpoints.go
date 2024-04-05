package handler

import (
	"crypto/rsa"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/handler/app"

	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/labstack/echo/v4"
)

// (GET /users/login)
func (s *Server) Login(ctx echo.Context) error {
	var cred generated.UserCredentials
	if err := ctx.Bind(&cred); err != nil {
		return err
	}

	usr, err := s.AuthService.Authenticate(cred.PhoneNumber, cred.Password)
	var authErr app.AuthenticationError
	if ok := errors.As(err, &authErr); ok {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err != nil {
		log.Printf("error authenticating user: %v", err)
		return err
	}

	tc := &TokenCreator{
		PrivateKey: _privateKey,
	}

	tokenString, err := tc.CreateAccessToken(usr)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, generated.LoginResponse{
		Id:          usr.ID(),
		AccessToken: tokenString,
	})
}

// Get my profile
// (GET /users/me)
func (s *Server) GetMyProfile(ctx echo.Context) error {
	userID, err := authenticatedUserID(ctx, _publicKey)
	if err != nil {
		return ctx.NoContent(http.StatusForbidden)
	}

	usr, err := s.UserService.GetProfile(userID)
	if err != nil {
		return err
	}

	if usr == nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, generated.UserProfile{
		Name:        usr.FullName(),
		PhoneNumber: usr.PhoneNumber(),
	})
}

// Update my profile
// (PUT /users/me)
func (s *Server) UpdateMyProfile(ctx echo.Context) error {
	userID, err := authenticatedUserID(ctx, _publicKey)
	if err != nil {
		return ctx.NoContent(http.StatusForbidden)
	}

	var profileForm generated.UserProfileForm
	if err := ctx.Bind(&profileForm); err != nil {
		return err
	}

	formErrs := validateUserProfileForm(profileForm)
	if len(formErrs) > 0 {
		return ctx.JSON(http.StatusBadRequest, formErrs)
	}

	err = s.UserService.UpdateProfile(userID, profileForm.FullName, profileForm.PhoneNumber)
	if errors.Is(err, app.ErrPhoneNumberAlreadyTaken) {
		return ctx.NoContent(http.StatusConflict)
	}

	if err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Register new user
// (POST /users/register)
func (s *Server) RegisterUser(ctx echo.Context) error {
	var regForm generated.UserRegistrationForm
	if err := ctx.Bind(&regForm); err != nil {
		return err
	}

	formErrs := validateRegistrationForm(regForm)
	if len(formErrs) > 0 {
		return ctx.JSON(http.StatusBadRequest, formErrs)
	}

	usr, err := s.UserService.RegisterUser(
		regForm.PhoneNumber,
		regForm.FullName,
		regForm.Password)

	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, generated.UserRegistrationResponse{
		Id: usr.ID(),
	})
}

func authenticatedUserID(ctx echo.Context, pubKey *rsa.PublicKey) (string, error) {
	authHeader := ctx.Request().Header.Get("Authorization")
	bearerToken, err := parseBearerToken(authHeader)
	if err != nil {
		return "", err
	}

	tv := &TokenVerifier{
		PublicKey: pubKey,
	}

	return tv.VerifyIdentify(bearerToken)
}

func parseBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New(("empty auth header"))
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid auth header")
	}

	return parts[1], nil
}

func validateRegistrationForm(form generated.UserRegistrationForm) []generated.FieldError {
	failures := make(map[string][]string)
	if !user.ValidPhoneNumberLength(form.PhoneNumber) {
		failures["phoneNumber"] = append(failures["phoneNumber"], errCodePhoneNumberLength)
	}

	if !user.ValidPhoneNumberPrefix(form.PhoneNumber) {
		failures["phoneNumber"] = append(failures["phoneNumber"], errCodePhoneNumberFormat)
	}

	if !user.ValidFullNameLength(form.FullName) {
		failures["fullName"] = append(failures["fullName"], errCodeFullNameLength)
	}

	if !user.ValidPasswordStrength(form.Password) {
		failures["password"] = append(failures["password"], errCodePasswordStrength)
	}

	var errors []generated.FieldError
	for field, codes := range failures {
		errors = append(errors, generated.FieldError{
			Name:  field,
			Codes: codes,
		})
	}

	return errors
}

func validateUserProfileForm(form generated.UserProfileForm) []generated.FieldError {
	failures := make(map[string][]string)

	if form.FullName != nil {
		if !user.ValidFullNameLength(*form.FullName) {
			failures["fullName"] = append(failures["fullName"], errCodeFullNameLength)
		}
	}

	if form.PhoneNumber != nil {
		if !user.ValidPhoneNumberLength(*form.PhoneNumber) {
			failures["phoneNumber"] = append(failures["phoneNumber"], errCodePhoneNumberLength)
		}

		if !user.ValidPhoneNumberPrefix(*form.PhoneNumber) {
			failures["phoneNumber"] = append(failures["phoneNumber"], errCodePhoneNumberFormat)
		}
	}

	var errors []generated.FieldError
	for field, codes := range failures {
		errors = append(errors, generated.FieldError{
			Name:  field,
			Codes: codes,
		})
	}

	return errors
}

const (
	errCodePhoneNumberLength = "PHONE_NUMBER_LENGTH"
	errCodePhoneNumberFormat = "PHONE_NUMBER_FORMAT"
	errCodeFullNameLength    = "FULL_NAME_LENGTH"
	errCodePasswordStrength  = "PASSWORD_STRENGTH"
)
