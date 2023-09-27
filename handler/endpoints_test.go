package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/dityuiri/UserServiceTest/common"
	"github.com/dityuiri/UserServiceTest/generated"
	"github.com/dityuiri/UserServiceTest/repository"
)

func initializeTestEchoServer(repo repository.RepositoryInterface) (generated.ServerInterface, *echo.Echo, *sync.WaitGroup) {
	e := echo.New()
	validate := validator.New()
	_ = validate.RegisterValidation("password", ValidatePassword)
	e.Validator = &UserRegistrationValidator{Validator: validate}
	var server generated.ServerInterface = &Server{JWTSecretKey: "key", Repository: repo}
	generated.RegisterHandlers(e, server)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = e.Start(":8080")
	}()

	return server, e, &wg
}

func generateNewToken(id string, key string) string {
	expirationTime := time.Now().Add(2 * time.Minute)
	claims := &jwt.MapClaims{
		"id":  id,
		"exp": jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(key))
	return tokenString
}

func TestUserRegister(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockRepository = repository.NewMockRepositoryInterface(mockCtrl)
	)

	sv, e, wg := initializeTestEchoServer(mockRepository)

	t.Run("all ok", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "Pass123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), repository.GetUserByPhoneNumberInput{PhoneNumber: "+62123456789"}).
			Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)
		mockRepository.EXPECT().InsertUser(gomock.Any(), gomock.Any()).Times(1)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		reqBody := `"invalid"`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("field validation rules violated", func(t *testing.T) {
		reqBody := `{"full_name": "Ha", "password": "hagasaurus", "phone_number": "+6780909080123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid password length", func(t *testing.T) {
		reqBody := `{"full_name": "Ha", "password": "ha", "phone_number": "+6780909080123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid password rule - no numeric", func(t *testing.T) {
		reqBody := `{"full_name": "Ha", "password": "Hagasaurus", "phone_number": "+6780909080123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid password rule - no special character", func(t *testing.T) {
		reqBody := `{"full_name": "Ha", "password": "Hagasaurus12", "phone_number": "+6780909080123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("get user phone number returns error", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "Pass123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), repository.GetUserByPhoneNumberInput{PhoneNumber: "+62123456789"}).
			Return(repository.GetUserByPhoneNumberOutput{}, errors.New("error")).Times(1)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "Pass123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), repository.GetUserByPhoneNumberInput{PhoneNumber: "+62123456789"}).
			Return(repository.GetUserByPhoneNumberOutput{}, nil).Times(1)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("insert user return error", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "Pass123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), repository.GetUserByPhoneNumberInput{PhoneNumber: "+62123456789"}).
			Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)
		mockRepository.EXPECT().InsertUser(gomock.Any(), gomock.Any()).Return(errors.New("error")).Times(1)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	_ = e.Shutdown(context.Background())
	wg.Wait()
}

func TestUserLogin(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockRepository = repository.NewMockRepositoryInterface(mockCtrl)

		knownHash, _ = bcrypt.GenerateFromPassword([]byte("correctPassword123!"), bcrypt.DefaultCost)

		userInput = repository.GetUserByPhoneNumberInput{
			PhoneNumber: "+62123456789",
		}

		userOutput = repository.GetUserByPhoneNumberOutput{
			Id:                   uuid.New(),
			Name:                 "Kurumi Ruru",
			Password:             string(knownHash),
			NumOfSuccessfulLogin: sql.NullInt32{Int32: 0},
		}
	)

	sv, e, wg := initializeTestEchoServer(mockRepository)

	t.Run("all ok", func(t *testing.T) {
		reqBody := `{"password": "correctPassword123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().UpsertUserLogin(gomock.Any(),
			repository.UpsertUserLoginInput{UserId: userOutput.Id,
				NumOfSuccessfulLogin: userOutput.NumOfSuccessfulLogin.Int32 + 1}).Return(nil).Times(1)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		reqBody := `{perkedel}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("field validation failed", func(t *testing.T) {
		reqBody := `{"password": "correctPassword123!"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("user not found", func(t *testing.T) {
		reqBody := `{"password": "correctPassword123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userInput).Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("get user by phone returning internal server error", func(t *testing.T) {
		reqBody := `{"password": "correctPassword123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userInput).
			Return(repository.GetUserByPhoneNumberOutput{}, errors.New("error")).Times(1)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("mismatched password", func(t *testing.T) {
		reqBody := `{"password": "haguUruna123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userInput).Return(userOutput, nil).Times(1)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("update user login returns error", func(t *testing.T) {
		reqBody := `{"password": "correctPassword123!", "phone_number": "+62123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().UpsertUserLogin(gomock.Any(),
			repository.UpsertUserLoginInput{UserId: userOutput.Id,
				NumOfSuccessfulLogin: userOutput.NumOfSuccessfulLogin.Int32 + 1}).Return(errors.New("error")).Times(1)

		if assert.NoError(t, sv.UserLogin(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	_ = e.Shutdown(context.Background())
	wg.Wait()
}

func TestGetUserProfile(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockRepository = repository.NewMockRepositoryInterface(mockCtrl)

		userId    = uuid.New()
		userInput = repository.GetUserByIdInput{
			Id: userId.String(),
		}

		userOutput = repository.GetUserByIdOutput{
			Id:          userId,
			Name:        "Kurumi Ruru",
			PhoneNumber: "628788889999",
		}
	)

	sv, e, wg := initializeTestEchoServer(mockRepository)

	t.Run("all ok", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, nil).Times(1)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("empty token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bear %s", "random"))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "random"))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("get user by id not found", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(repository.GetUserByIdOutput{}, common.ErrUserNotFound).Times(1)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("get user by id return error", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(repository.GetUserByIdOutput{}, errors.New("error")).Times(1)

		if assert.NoError(t, sv.GetUserProfile(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	_ = e.Shutdown(context.Background())
	wg.Wait()
}

func TestUpdateUserProfile(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockRepository = repository.NewMockRepositoryInterface(mockCtrl)

		userId    = uuid.New()
		userInput = repository.GetUserByIdInput{
			Id: userId.String(),
		}

		userOutput = repository.GetUserByIdOutput{
			Id:          userId,
			Name:        "Kurumi Ruru",
			PhoneNumber: "628788889999",
		}
	)

	sv, e, wg := initializeTestEchoServer(mockRepository)

	t.Run("all ok", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		updateUserInput := repository.UpdateUserInput{
			Id:          userOutput.Id.String(),
			Name:        "Mirapa Ruru",
			PhoneNumber: "+6212345678219",
		}

		userPhoneInput := repository.GetUserByPhoneNumberInput{
			PhoneNumber: updateUserInput.PhoneNumber,
		}

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userPhoneInput).Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)
		mockRepository.EXPECT().UpdateUser(gomock.Any(), updateUserInput).Return(nil).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("empty token", func(t *testing.T) {
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid header format", func(t *testing.T) {
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", "random"))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"invalid"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("empty request body", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("user not found", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, common.ErrUserNotFound).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("get user by id returning error", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, errors.New("error")).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("no changes", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		updateUserInput := repository.UpdateUserInput{
			Id:          userOutput.Id.String(),
			Name:        "Mirapa Ruru",
			PhoneNumber: "+6212345678219",
		}

		noChangesUserOutput := repository.GetUserByIdOutput{
			Id:          userId,
			Name:        updateUserInput.Name,
			PhoneNumber: updateUserInput.PhoneNumber,
		}

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(noChangesUserOutput, nil).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusNoContent, rec.Code)
		}
	})

	t.Run("get user by phone number return error", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		updateUserInput := repository.UpdateUserInput{
			Id:          userOutput.Id.String(),
			Name:        "Mirapa Ruru",
			PhoneNumber: "+6212345678219",
		}

		userPhoneInput := repository.GetUserByPhoneNumberInput{
			PhoneNumber: updateUserInput.PhoneNumber,
		}

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userPhoneInput).Return(repository.GetUserByPhoneNumberOutput{}, errors.New("error")).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("existing user with phone number found", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		updateUserInput := repository.UpdateUserInput{
			Id:          userOutput.Id.String(),
			Name:        "Mirapa Ruru",
			PhoneNumber: "+6212345678219",
		}

		userPhoneInput := repository.GetUserByPhoneNumberInput{
			PhoneNumber: updateUserInput.PhoneNumber,
		}

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userPhoneInput).Return(repository.GetUserByPhoneNumberOutput{Name: "Haga Uruna"}, nil).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("update user returning error", func(t *testing.T) {
		generatedToken := generateNewToken(userId.String(), "key")
		reqBody := `{"full_name": "Mirapa Ruru", "phone_number": "+6212345678219"}`
		req := httptest.NewRequest(http.MethodPatch, "/user/profile", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", generatedToken))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		updateUserInput := repository.UpdateUserInput{
			Id:          userOutput.Id.String(),
			Name:        "Mirapa Ruru",
			PhoneNumber: "+6212345678219",
		}

		userPhoneInput := repository.GetUserByPhoneNumberInput{
			PhoneNumber: updateUserInput.PhoneNumber,
		}

		mockRepository.EXPECT().GetUserById(gomock.Any(), userInput).Return(userOutput, nil).Times(1)
		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), userPhoneInput).Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)
		mockRepository.EXPECT().UpdateUser(gomock.Any(), updateUserInput).Return(errors.New("error")).Times(1)

		if assert.NoError(t, sv.UpdateUserProfile(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	_ = e.Shutdown(context.Background())
	wg.Wait()
}
