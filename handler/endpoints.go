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
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
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
				Password:    string(hashedPassword),
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

// UserLogin : POST /user/login
func (s *Server) UserLogin(ctx echo.Context) error {
	var (
		req  generated.UserLoginRequest
		resp generated.UserLoginResponse

		standardCtx = ctx.Request().Context()
	)

	// Retrieve request body
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	// Required field validation
	err := ctx.Validate(req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: "PhoneNumber and Password are mandatory",
		})
	}

	// Get user to compare the password
	getUserInput := repository.GetUserByPhoneNumberInput{PhoneNumber: req.PhoneNumber}
	user, err := s.Repository.GetUserByPhoneNumber(standardCtx, getUserInput)
	if err != nil {
		if err == common.ErrUserNotFound {
			// Case when user not found
			return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
				Message: err.Error(),
			})
		}

		ctx.Logger().Errorf("GetUserLogin error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Compare supplied password with the user password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
				Message: "Mismatched password",
			})
		}

		ctx.Logger().Errorf("CompareHashAndPassword error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Generate JWT token
	token, err := s.generateJWTToken(user.Id.String())
	if err != nil {
		ctx.Logger().Errorf("generateJWTToken error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Increment successful login
	updateUserLoginInput := repository.UpsertUserLoginInput{
		UserId:               user.Id,
		NumOfSuccessfulLogin: user.NumOfSuccessfulLogin.Int32 + 1,
	}

	err = s.Repository.UpsertUserLogin(standardCtx, updateUserLoginInput)
	if err != nil {
		ctx.Logger().Errorf("UpdateUserLogin error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	resp.Id = user.Id.String()
	resp.Token = token
	return ctx.JSON(http.StatusOK, resp)
}
