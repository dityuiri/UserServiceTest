package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/dityuiri/UserServiceTest/common"
	"github.com/dityuiri/UserServiceTest/generated"
	"github.com/dityuiri/UserServiceTest/repository"
)

func initializeTestEchoServer(repo repository.RepositoryInterface) (generated.ServerInterface, *echo.Echo, *sync.WaitGroup) {
	e := echo.New()
	var server generated.ServerInterface = &Server{Repository: repo}
	generated.RegisterHandlers(e, server)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = e.Start(":8080")
	}()

	return server, e, &wg
}

func TestUserRegister(t *testing.T) {
	var (
		mockCtrl       = gomock.NewController(t)
		mockRepository = repository.NewMockRepositoryInterface(mockCtrl)
	)

	sv, e, wg := initializeTestEchoServer(mockRepository)

	t.Run("all ok", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "pass123", "phone_number": "+62123456789"}`
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

	t.Run("get user phone number returns error", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "pass123", "phone_number": "+62123456789"}`
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
		reqBody := `{"full_name": "Haga Uruna", "password": "pass123", "phone_number": "+62123456789"}`
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

	t.Run("hash password returns error", func(t *testing.T) {
		invalidPass := strings.Repeat("A", 73)
		reqBody := fmt.Sprintf(`{"full_name": "Haga Uruna", "password": "%s", "phone_number": "+62123456789"}`, invalidPass)
		req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockRepository.EXPECT().GetUserByPhoneNumber(gomock.Any(), repository.GetUserByPhoneNumberInput{PhoneNumber: "+62123456789"}).
			Return(repository.GetUserByPhoneNumberOutput{}, common.ErrUserNotFound).Times(1)

		if assert.NoError(t, sv.UserRegister(c)) {
			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotEmpty(t, rec.Body.String())
		}
	})

	t.Run("insert user return error", func(t *testing.T) {
		reqBody := `{"full_name": "Haga Uruna", "password": "pass123", "phone_number": "+62123456789"}`
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
