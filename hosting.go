package main

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"github.com/robertkrimen/otto"
	"net/http"
)

var E *echo.Echo

func StartServer() {
	E = echo.New()
	E.Any("/dispatch/:project/:event", EventDispatcher)
	E.GET("/system/projects", ReadProjectsHandler)
	E.POST("/system/projects", ApplyProjectHandler)
	E.DELETE("/system/projects/:name", DeleteProjectHandler)
	err := E.Start(":3000")
	if err != nil {
		panic(err)
	}
}

func EventDispatcher(ctx echo.Context) error {
	var project Project
	err := Database.C("projects").Find(map[string]string{
		"name": ctx.Param("project"),
	}).One(&project)
	if err != nil {
		return errors.New("project not found")
	}
	actions := map[string]otto.Value{}
	vm := otto.New()
	err = vm.Set("endpoint", func(call otto.FunctionCall) otto.Value {
		name, _ := call.Argument(0).ToString()
		actions[name] = call.Argument(1)
		return otto.Value{}
	})
	if err != nil {
		fmt.Println(err)
	}
	_, err = vm.Run(project.Definition)
	if err != nil {
		fmt.Println(err)
	}
	result, err := actions[ctx.Param("event")].Call(otto.TrueValue())
	if err != nil {
		fmt.Println(err)
	}
	out, _ := result.ToString()
	return ctx.JSON(http.StatusOK, out)
}
