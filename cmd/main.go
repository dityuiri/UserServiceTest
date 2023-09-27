package main

import (
	"os"

	"github.com/dityuiri/UserServiceTest/generated"
	"github.com/dityuiri/UserServiceTest/handler"
	"github.com/dityuiri/UserServiceTest/repository"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Validator = &handler.UserRegistrationValidator{Validator: setupValidator()}

	var server generated.ServerInterface = newServer()
	generated.RegisterHandlers(e, server)
	e.Logger.Fatal(e.Start(":1323"))
}

func newServer() *handler.Server {
	dbDsn := os.Getenv("DATABASE_URL")
	var repo repository.RepositoryInterface = repository.NewRepository(repository.NewRepositoryOptions{
		Dsn: dbDsn,
	})
	opts := handler.NewServerOptions{
		JWTSecretKey: os.Getenv("JWT_SECRET_KEY"),
		Repository:   repo,
	}
	return handler.NewServer(opts)
}

func setupValidator() *validator.Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("password", handler.ValidatePassword)

	return validate
}
