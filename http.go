package main

import (
	"github.com/labstack/echo"
	"gopkg.in/go-playground/validator.v9"
)

type response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type (
	appValidator struct {
		validator *validator.Validate
	}
)

func (cv *appValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var rest *echo.Echo

func bootstrapRestServer() {
	rest = echo.New()
	rest.Validator = &appValidator{validator: validator.New()}
	handleRestRoutes()
}

func handleRestRoutes() {
	rest.POST("/auth/session", loginController)
	rest.POST("/auth/user", registerController)
	rest.POST("/system/projects", createProjectController)
	rest.GET("/system/projects", readProjectsController)
	rest.PUT("/system/projects/:id", updateProjectController)
	rest.DELETE("/system/projects/:id", deleteProjectController)
	rest.POST("/system/projects/:project/databases", createDatabaseController)
	rest.GET("/system/projects/:project/databases", readDatabasesController)
	rest.PUT("/system/projects/:project/databases/:id", updateDatabaseController)
	rest.DELETE("/system/projects/:project/databases/:id", deleteDatabaseController)

	rest.POST("/apps/:database/:collection", createResourceController)
	rest.GET("/apps/:database/:collection", readResourcesController)
}
