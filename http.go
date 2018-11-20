package main

import (
	"github.com/labstack/echo"
	"gopkg.in/go-playground/validator.v9"
)

type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type (
	AppValidator struct {
		validator *validator.Validate
	}
)

func (cv *AppValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var rest *echo.Echo

func BootstrapRestServer() {
	rest = echo.New()
	rest.Validator = &AppValidator{validator: validator.New()}
	HandleRestRoutes()
}

func HandleRestRoutes() {
	rest.POST("/auth/session", LoginController)
	rest.POST("/auth/user", RegisterController)
	rest.POST("/system/projects", CreateProjectController)
	rest.GET("/system/projects", ReadProjectsController)
	rest.PUT("/system/projects/:id", UpdateProjectController)
	rest.DELETE("/system/projects/:id", DeleteProjectController)
}
