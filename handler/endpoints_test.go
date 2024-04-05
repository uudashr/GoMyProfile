package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/handler/model/user"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
)

type fixture struct {
	ctrl     *gomock.Controller
	userRepo *repository.MockRepositoryInterface
	svr      *Server
}

func setup(t *testing.T) *fixture {
	ctrl := gomock.NewController(t)
	userRepo := repository.NewMockRepositoryInterface(ctrl)
	svr := NewServer(NewServerOptions{
		Repository: userRepo,
	})

	return &fixture{
		ctrl:     ctrl,
		userRepo: userRepo,
		svr:      svr,
	}
}

func (f *fixture) tearDown() {
	f.ctrl.Finish()
}

func TestRegisterUser(t *testing.T) {
	testCases := map[string]struct {
		regForm             generated.UserRegistrationForm
		expectStatusCode    int
		expectContainsError map[string][]string
	}{
		"success": {
			regForm: generated.UserRegistrationForm{
				PhoneNumber: "+628174546647",
				FullName:    "John Doe",
				Password:    "Secret123!",
			},
			expectStatusCode: http.StatusOK,
		},
		"phoneNumber too short": {
			regForm: generated.UserRegistrationForm{
				PhoneNumber: "+62817454",
				FullName:    "John Doe",
				Password:    "Secret123!",
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"phoneNumber": {"PHONE_NUMBER_LENGTH"},
			},
		},
		"phoneNumber invalid format": {
			regForm: generated.UserRegistrationForm{
				PhoneNumber: "+618174546647",
				FullName:    "John Doe",
				Password:    "Secret123!",
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"phoneNumber": {"PHONE_NUMBER_FORMAT"},
			},
		},
		"fullName too long": {
			regForm: generated.UserRegistrationForm{
				PhoneNumber: "+628174546647",
				FullName:    "John Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe Doe",
				Password:    "Secret123!",
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"fullName": {"FULL_NAME_LENGTH"},
			},
		},
		"password not strong enough": {
			regForm: generated.UserRegistrationForm{
				PhoneNumber: "+628174546647",
				FullName:    "John Doe",
				Password:    "weak",
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"password": {"PASSWORD_STRENGTH"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			fix := setup(t)
			defer fix.tearDown()

			// When
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.regForm); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(buf.Bytes()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			var newlyStoredUser *user.User
			if tc.expectStatusCode == http.StatusOK {
				fix.userRepo.EXPECT().Store(gomock.Any()).DoAndReturn(func(u *user.User) (*user.User, error) {
					if u.PhoneNumber() != tc.regForm.PhoneNumber {
						return nil, errors.New("phoneNumber is not equal")
					}

					if u.FullName() != tc.regForm.FullName {
						return nil, errors.New("fullName is not equal")
					}

					newlyStoredUser = u

					return u, nil
				})
			}

			c := echo.New().NewContext(req, rec)
			err := fix.svr.RegisterUser(c)

			// Then
			if err != nil {
				t.Fatal(err)
			}

			if got, want := rec.Code, tc.expectStatusCode; got != want {
				t.Fatalf("stausCode got %d, want %d", got, want)
			}

			if rec.Code == http.StatusOK {
				var res generated.UserRegistrationResponse
				if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
					t.Fatal(err)
				}

				if got, want := res.Id, newlyStoredUser.ID(); got != want {
					t.Fatalf("id got %s, want %s", got, want)
				}
			}

			if rec.Code == http.StatusBadRequest {
				var resErrs []generated.FieldError
				if err := json.NewDecoder(rec.Body).Decode(&resErrs); err != nil {
					t.Fatal(err)
				}

				for fieldName, errCodes := range tc.expectContainsError {
					for _, code := range errCodes {
						found := false
						for _, e := range resErrs {
							if e.Name == fieldName && slices.Contains(e.Codes, code) {
								found = true
								break
							}
						}

						if !found {
							t.Fatalf("FieldError not found: want fieldName=%s code=%s, got errors=%+v", fieldName, code, resErrs)
						}
					}
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	usrPassword := "Secret123!"
	user1, err := user.NewWithPassword("jdoe", "+628174546647", "John Doe", usrPassword)
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		creds            generated.UserCredentials
		returnedUser     *user.User
		expectStatusCode int
	}{
		"success": {
			creds: generated.UserCredentials{
				PhoneNumber: user1.PhoneNumber(),
				Password:    usrPassword,
			},
			returnedUser:     user1,
			expectStatusCode: http.StatusOK,
		},
		"phone number not found": {
			creds: generated.UserCredentials{
				PhoneNumber: "+628174546648",
				Password:    usrPassword,
			},
			returnedUser:     nil,
			expectStatusCode: http.StatusBadRequest,
		},
		"invalid password": {
			creds: generated.UserCredentials{
				PhoneNumber: user1.PhoneNumber(),
				Password:    usrPassword + "x",
			},
			returnedUser:     user1,
			expectStatusCode: http.StatusBadRequest,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			fix := setup(t)
			defer fix.tearDown()

			// When
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.creds); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(buf.Bytes()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			fix.userRepo.EXPECT().GetByPhoneNumber(tc.creds.PhoneNumber).Return(tc.returnedUser, nil)

			c := echo.New().NewContext(req, rec)
			err = fix.svr.Login(c)

			// Then
			if err != nil {
				t.Fatal(err)
			}

			if got, want := rec.Code, tc.expectStatusCode; got != want {
				t.Fatalf("statusCode got %d, want %d", got, want)
			}

			if tc.expectStatusCode != http.StatusOK {
				return
			}

			var res generated.LoginResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatal(err)
			}

			if got, want := res.Id, user1.ID(); got != want {
				t.Fatalf("id got %s, want %s", got, want)
			}

			if got := res.AccessToken; got == "" {
				t.Fatalf("accessToken got %s, want not empty", got)
			}
		})
	}
}

func TestGetMyProfile(t *testing.T) {
	user1, err := user.NewWithPassword("jdoe", "+628174546647", "John Doe", "Secret123!")
	if err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		user             *user.User
		tokenFn          func(*user.User) (string, error)
		invalidToken     bool
		expectStatusCode int
	}{
		"success": {
			user: user1,
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusOK,
		},
		"invalid token": {
			user: user1,
			tokenFn: func(u *user.User) (string, error) {
				return "invalid token", nil
			},
			invalidToken:     true,
			expectStatusCode: http.StatusForbidden,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			fix := setup(t)
			defer fix.tearDown()

			// When
			accessToken, err := tc.tokenFn(tc.user)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
			rec := httptest.NewRecorder()

			if !tc.invalidToken {
				fix.userRepo.EXPECT().GetByID(tc.user.ID()).Return(tc.user, nil)
			}

			c := echo.New().NewContext(req, rec)
			err = fix.svr.GetMyProfile(c)

			// Then
			if err != nil {
				t.Fatal(err)
			}

			if got, want := rec.Code, tc.expectStatusCode; got != want {
				t.Fatalf("statusCode got %d, want %d", got, want)
			}

			if rec.Code != http.StatusOK {
				return
			}

			var res generated.UserProfile
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatal(err)
			}

			if got, want := res.Name, tc.user.FullName(); got != want {
				t.Fatalf("name got %s, want %s", got, want)
			}

			if got, want := res.PhoneNumber, tc.user.PhoneNumber(); got != want {
				t.Fatalf("phoneNumber got %s, want %s", got, want)
			}
		})
	}
}

func TestUpdateMyProfile(t *testing.T) {
	testCases := map[string]struct {
		phoneNumber         string
		fullName            string
		profileForm         generated.UserProfileForm
		tokenFn             func(*user.User) (string, error)
		noUpdate            bool
		invalidToken        bool
		expectStatusCode    int
		expectContainsError map[string][]string

		skip bool
	}{
		"update both": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+628174546648"),
				FullName:    strPtr("John Wick"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusNoContent,
		},
		"update phone only": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+628174546648"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusNoContent,
		},
		"update name only": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				FullName: strPtr("John Wick"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusNoContent,
		},
		"no details sent": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			noUpdate:         true,
			expectStatusCode: http.StatusNoContent,
		},
		"no changes": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+628174546647"),
				FullName:    strPtr("John Doe"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			noUpdate:         true,
			expectStatusCode: http.StatusNoContent,
		},
		"invalid name length": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+628174546648"),
				FullName:    strPtr("Jo"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"fullName": {"FULL_NAME_LENGTH"},
			},
		},
		"invalid phone format": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+618174546648"),
				FullName:    strPtr("John Wick"),
			},
			tokenFn: func(u *user.User) (string, error) {
				tc := &TokenCreator{
					PrivateKey: _privateKey,
				}
				return tc.CreateAccessToken(u)
			},
			expectStatusCode: http.StatusBadRequest,
			expectContainsError: map[string][]string{
				"phoneNumber": {"PHONE_NUMBER_FORMAT"},
			},
		},
		"invalid token": {
			phoneNumber: "+628174546647",
			fullName:    "John Doe",
			profileForm: generated.UserProfileForm{
				PhoneNumber: strPtr("+628174546648"),
				FullName:    strPtr("John Wick"),
			},
			tokenFn: func(u *user.User) (string, error) {
				return "invalid token", nil
			},
			invalidToken:     true,
			expectStatusCode: http.StatusForbidden,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}

			// Given
			fix := setup(t)
			defer fix.tearDown()

			storedUser, err := user.NewWithPassword("jdoe", tc.phoneNumber, tc.fullName, "Secret123!")
			if err != nil {
				t.Fatal(err)
			}

			// When
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.profileForm); err != nil {
				t.Fatal(err)
			}

			accessToken, err := tc.tokenFn(storedUser)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(buf.Bytes()))
			req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", accessToken))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			if !tc.invalidToken && len(tc.expectContainsError) == 0 {
				fix.userRepo.EXPECT().GetByID(storedUser.ID()).Return(storedUser, nil)

				if tc.profileForm.PhoneNumber != nil && *tc.profileForm.PhoneNumber != tc.phoneNumber {
					fix.userRepo.EXPECT().GetByPhoneNumber(*tc.profileForm.PhoneNumber).Return(nil, nil)
				}

				if !tc.noUpdate {
					fix.userRepo.EXPECT().Update(storedUser).Return(nil)
				}
			}

			c := echo.New().NewContext(req, rec)
			err = fix.svr.UpdateMyProfile(c)

			// Then
			if err != nil {
				t.Fatal(err)
			}

			if got, want := rec.Code, tc.expectStatusCode; got != want {
				t.Fatalf("statusCode got %d, want %d", got, want)
			}

			if rec.Code == http.StatusBadRequest {
				var resErrs []generated.FieldError
				if err := json.NewDecoder(rec.Body).Decode(&resErrs); err != nil {
					t.Fatal(err)
				}

				for fieldName, errCodes := range tc.expectContainsError {
					for _, code := range errCodes {
						found := false
						for _, e := range resErrs {
							if e.Name == fieldName && slices.Contains(e.Codes, code) {
								found = true
								break
							}
						}

						if !found {
							t.Fatalf("FieldError not found: want fieldName=%s code=%s, got errors=%+v", fieldName, code, resErrs)
						}
					}
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
