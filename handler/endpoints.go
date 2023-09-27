package handler

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/dityuiri/UserServiceTest/common"
	"github.com/dityuiri/UserServiceTest/generated"
	"github.com/dityuiri/UserServiceTest/repository"
)

// UserRegister : POST /user/register
func (s *Server) UserRegister(ctx echo.Context) error {
	var (
		req  generated.UserRegisterRequest
		resp generated.UserRegisterCreatedResponse

		standardCtx = ctx.Request().Context()
	)

	// Retrieve request body
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.MultipleErrorResponse{
			Messages: []string{"Invalid request body"},
		})
	}

	// Field validation
	err := ctx.Validate(req)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errMessages := TranslateErrorMessages(validationErrors)
			return ctx.JSON(http.StatusBadRequest, generated.MultipleErrorResponse{
				Messages: errMessages,
			})
		}
	}

	// Validate if user already created
	getUserInput := repository.GetUserByPhoneNumberInput{PhoneNumber: req.PhoneNumber}
	_, err = s.Repository.GetUserByPhoneNumber(standardCtx, getUserInput)
	if err != nil {
		if err != common.ErrUserNotFound {
			ctx.Logger().Errorf("GetUserByPhoneNumber error: %s", err.Error())
			return ctx.JSON(http.StatusInternalServerError, generated.MultipleErrorResponse{
				Messages: []string{err.Error()},
			})
		}

		// Normal case is when user isn't exist in the database
		if err == common.ErrUserNotFound {
			// Hash and Salt the password
			hashedPassword, err := s.hashPassword(req.Password)
			if err != nil {
				ctx.Logger().Errorf("hashPassword error: %s", err.Error())
				return ctx.JSON(http.StatusInternalServerError, generated.MultipleErrorResponse{
					Messages: []string{err.Error()},
				})
			}

			// Insert user
			insertUserInput := repository.InsertUserInput{
				Id:          uuid.New(),
				PhoneNumber: req.PhoneNumber,
				Name:        req.FullName,
				Password:    hashedPassword,
			}

			err = s.Repository.InsertUser(standardCtx, insertUserInput)
			if err != nil {
				ctx.Logger().Errorf("InsertUser error: %s", err.Error())
				return ctx.JSON(http.StatusInternalServerError, generated.MultipleErrorResponse{
					Messages: []string{err.Error()},
				})
			}

			// Success response
			resp.Id = insertUserInput.Id.String()
			return ctx.JSON(http.StatusCreated, resp)
		}
	}

	// Return 422 if user already created
	return ctx.JSON(http.StatusUnprocessableEntity, generated.MultipleErrorResponse{
		Messages: []string{"User already exists"},
	})
}

func (s *Server) hashPassword(pass string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
