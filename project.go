package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"net/http"
)

type Project struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

func createProjectController(c echo.Context) error {
	user, err := getUserId(c)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	project := new(Project)
	err = c.Bind(&project)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	project.Author = user["_id"].(string)
	if err = c.Validate(project); err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
		})
		return nil
	}
	err = applicationDatabase.C("projects").Insert(project)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot insert project to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Project successfully inserted",
		Data:    project,
	})
	return nil
}

func readProjectsController(c echo.Context) error {
	user, err := getUserId(c)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	var projects []Project
	err = applicationDatabase.C("projects").Find(echo.Map{
		"author": user["_id"],
	}).All(&projects)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot query projects",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Successfully queried projects",
		Data:    projects,
	})
	return nil
}

func updateProjectController(c echo.Context) error {
	user, err := getUserId(c)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	project := new(Project)
	err = c.Bind(&project)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	project.Author = user["_id"].(string)
	if err = c.Validate(project); err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	query["author"] = user["_id"]
	err = applicationDatabase.C("projects").Update(query, project)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot update project to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Project successfully updated",
		Data:    project,
	})
	return nil
}

func deleteProjectController(c echo.Context) error {
	user, err := getUserId(c)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	query["author"] = user["_id"]
	err = applicationDatabase.C("projects").Remove(query)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot delete project to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Project successfully delete",
		Data:    nil,
	})
	return nil
}
