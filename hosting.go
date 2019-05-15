package main

import "github.com/labstack/echo"

var E *echo.Echo

func StartServer() {
	E = echo.New()
	E.GET("/system/projects", ReadProjectsHandler)
	E.POST("/system/projects", ApplyProjectHandler)
	E.DELETE("/system/projects/:name", DeleteProjectHandler)
	err := E.Start(":3000")
	if err != nil {
		panic(err)
	}
}
