package main

import (
	"errors"
	"github.com/labstack/echo"
	"gopkg.in/yaml.v2"
	"strings"
)

var E *echo.Echo

func StartServer() {
	E = echo.New()
	E.GET("/services/:project/:event", DispatchEvent)
	E.GET("/system/projects", ReadProjectsHandler)
	E.POST("/system/projects", ApplyProjectHandler)
	E.DELETE("/system/projects/:name", DeleteProjectHandler)
	err := E.Start(":3000")
	if err != nil {
		panic(err)
	}
}

type Event struct {
	Fields   map[string]string `yaml:"fields"`
	Handlers []string          `yaml:"handlers"`
}
type Definition struct {
	Events map[string]Event `yaml:"events"`
}

func DispatchEvent(ctx echo.Context) error {
	var project Project
	err := Database.C("projects").Find(map[string]string{
		"name": ctx.Param("project"),
	}).One(&project)
	if err != nil {
		return errors.New("cannot find project")
	}
	def := Definition{}
	eventName := ctx.Param("event")
	err = yaml.Unmarshal([]byte(project.Definition), &def)
	if err != nil {
		return errors.New("server error: cannot parse project definition")
	}
	if _, ok := def.Events[eventName]; ok {
		event := def.Events[eventName]
		errored := false
		erroredFieldName := ""
		var body interface{}
		_ = ctx.Bind(&body)
		for name, field := range event.Fields {
			if strings.Contains(field, "required") {
				if _, ok := body.(map[string]interface{})[name]; !ok {
					errored = true
					erroredFieldName = name
				}
			}
		}
		if errored {
			return errors.New("field: " + erroredFieldName)
		}
		for _, handler := range event.Handlers {
			if strings.HasPrefix(handler, "db.insert") {
			
			} else if strings.HasPrefix(handler, "db.delete") {

			} else if strings.HasPrefix(handler, "db.set") {

			}
		}
	} else {
		return errors.New("cannot dispatch given event")
	}
}
