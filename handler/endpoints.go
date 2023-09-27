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

		ctx.Logger().Errorf("GetUserByPhoneNumber error: %s", err.Error())
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

// GetUserProfile : GET /user/profile
func (s *Server) GetUserProfile(ctx echo.Context) error {
	var (
		resp        generated.GetUserProfileResponse
		standardCtx = ctx.Request().Context()
	)

	// Retrieve and Get ID from JWT Token
	userId, err := s.retrieveAndGetIdFromJWTToken(ctx)
	if err != nil {
		return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Get user profile
	getUserInput := repository.GetUserByIdInput{Id: userId}
	user, err := s.Repository.GetUserById(standardCtx, getUserInput)
	if err != nil {
		if err == common.ErrUserNotFound {
			// Follow the specification to return it as 403
			return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{
				Message: err.Error(),
			})
		}

		ctx.Logger().Errorf("GetUserById error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	resp.Name = user.Name
	resp.PhoneNumber = user.PhoneNumber
	return ctx.JSON(http.StatusOK, resp)
}

// UpdateUserProfile : PATCH /user/profile
func (s *Server) UpdateUserProfile(ctx echo.Context) error {
	var (
		req         generated.UpdateUserProfileRequest
		resp        generated.SuccessMessageResponse
		standardCtx = ctx.Request().Context()
	)

	// Retrieve and Get ID from JWT Token
	userId, err := s.retrieveAndGetIdFromJWTToken(ctx)
	if err != nil {
		return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Retrieve request body
	if err = ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	// Check if both of the fields are empty
	if req.PhoneNumber == nil && req.FullName == nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
			Message: "Empty request body",
		})
	}

	// Get user by id to get current profile of the user
	getUserInput := repository.GetUserByIdInput{Id: userId}
	user, err := s.Repository.GetUserById(standardCtx, getUserInput)
	if err != nil {
		if err == common.ErrUserNotFound {
			// Follow the specification to return it as 403
			return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{
				Message: err.Error(),
			})
		}

		ctx.Logger().Errorf("GetUserById error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	// Pre-fill input for update user with existing profile
	updateUserInput := repository.UpdateUserInput{
		Id:          userId,
		PhoneNumber: user.PhoneNumber,
		Name:        user.Name,
	}

	// Check if any changes happen to the current one
	var isPhoneChanged, isNameChanged bool
	if req.PhoneNumber != nil {
		if user.PhoneNumber != *req.PhoneNumber {
			isPhoneChanged = true
		}

		updateUserInput.PhoneNumber = *req.PhoneNumber
	}

	if req.FullName != nil {
		if user.Name != *req.FullName {
			isNameChanged = true
			updateUserInput.Name = *req.FullName
		}
	}

	// Return no content if no changes happened
	if !isPhoneChanged && !isNameChanged {
		return ctx.JSON(http.StatusNoContent, nil)
	}

	// If phone number changed, check for existing user
	if isPhoneChanged {
		getUserByPhoneInput := repository.GetUserByPhoneNumberInput{PhoneNumber: *req.PhoneNumber}
		existingUser, err := s.Repository.GetUserByPhoneNumber(standardCtx, getUserByPhoneInput)
		// Return for other errors
		if err != nil && err != common.ErrUserNotFound {
			ctx.Logger().Errorf("GetUserByPhoneNumber error: %s", err.Error())
			return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
				Message: err.Error(),
			})
		}

		// Return conflict status code
		if err == nil && existingUser.Name != "" {
			return ctx.JSON(http.StatusConflict, generated.ErrorResponse{
				Message: "phone number exists",
			})
		}
	}

	// Continue the update process
	err = s.Repository.UpdateUser(standardCtx, updateUserInput)
	if err != nil {
		ctx.Logger().Errorf("UpdateUser error: %s", err.Error())
		return ctx.JSON(http.StatusInternalServerError, generated.ErrorResponse{
			Message: err.Error(),
		})
	}

	resp.Message = "changes applied successfully"
	return ctx.JSON(http.StatusOK, resp)
}
