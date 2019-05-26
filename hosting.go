package main

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/labstack/echo"
	"github.com/robertkrimen/otto"
	"net/http"
	"time"
)

var connections map[string]*mgo.Database
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
	_ = vm.Set("read", ReadDatabaseFunction)
	_, err = vm.Run("var action = function action(name, func) {\r\n  endpoint(name, function (ctx) {\r\n    return JSON.stringify(func(ctx));\r\n  });\r\n};" + project.Definition)
	if err != nil {
		fmt.Println(err)
	}
	var body interface{}
	_ = ctx.Bind(&body)
	result, err := actions[ctx.Param("event")].Call(otto.TrueValue(), echo.Map{
		"body":    body,
		"headers": ctx.Request().Header,
		"client": echo.Map{
			"ip": ctx.RealIP(),
		},
		"now": time.Now().Unix(),
	})
	if err != nil {
		fmt.Println(err)
	}
	out, _ := result.ToString()
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	return ctx.String(http.StatusOK, out)
}

func ReadDatabaseFunction(project Project) func(call otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		conn := EnsureConnectionExists(project)
		var result []interface{}
		query := echo.Map{}
		for _, k := range call.Argument(1).Object().Keys() {
			query[k], _ = call.Argument(1).Object().Get(k)
		}
		var limit int64
		var skip int64
		limitVal, err := call.Argument(2).Object().Get("limit")
		if err != nil {
			limit = 50
		} else {
			limit, _ = limitVal.ToInteger()
		}
		skipVal, err := call.Argument(2).Object().Get("skip")
		if err != nil {
			skip = 0
		} else {
			skip, _ = skipVal.ToInteger()
		}
		_ = conn.C(call.Argument(0).String()).Find(query).Limit(int(limit)).Skip(int(skip)).All(&result)
		out, _ := otto.ToValue(result)
		return out
	}
}

func EnsureConnectionExists(project Project) *mgo.Database {
	if _, ok := connections[project.Name]; !ok {
		ses, _ := mgo.Dial(project.Mongo)
		connections[project.Name] = ses.DB("")
	}
	return connections[project.Name]
}
