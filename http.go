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
	rest.POST("/auth/session", loginController)
	rest.POST("/auth/user", registerController)
	rest.POST("/system/projects", createProjectController)
	rest.GET("/system/projects", readProjectsController)
	rest.PUT("/system/projects/:id", updateProjectController)
	rest.DELETE("/system/projects/:id", deleteProjectController)
}
