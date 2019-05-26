package main

import (
	"errors"
	"github.com/labstack/echo"
	"net/http"
)

type Project struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
	Mongo      string `json:"mongo"`
}

func ReadProjectsHandler(ctx echo.Context) error {
	projects := make([]Project, 0)
	err := Database.C("projects").Find(nil).All(&projects)
	if err != nil {
		return errors.New("cannot index projects")
	}
	return ctx.JSON(http.StatusOK, projects)
}

func ApplyProjectHandler(ctx echo.Context) error {
	var body Project
	err := ctx.Bind(&body)
	if err != nil {
		return err
	}
	var project Project
	err = Database.C("projects").Find(map[string]string{
		"name": body.Name,
	}).One(&project)
	if err != nil {
		err = Database.C("projects").Insert(&body)
		if err != nil {
			return errors.New("cannot insert project")
		}
	} else {
		project.Definition = body.Definition
		project.Mongo = body.Mongo
		err := Database.C("projects").Update(map[string]string{
			"name": body.Name,
		}, project)
		if err != nil {
			return errors.New("cannot update project")
		}
	}
	return ctx.JSON(http.StatusOK, body)
}

func DeleteProjectHandler(ctx echo.Context) error {
	err := Database.C("projects").Remove(map[string]string{
		"name": ctx.Param("name"),
	})
	if err != nil {
		return errors.New("cannot delete project")
	}
	return ctx.JSON(http.StatusOK, map[string]string{})
}
