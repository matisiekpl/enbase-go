package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"net/http"
)

type database struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Author      string        `json:"author"`
	Project     string        `json:"project"`
	Rules       echo.Map      `json:"rules"`
	Reads       int           `json:"reads"`
	Creates     int           `json:"writes"`
	Updates     int           `json:"updates"`
	Deletes     int           `json:"deletes"`
}

func createDatabaseController(c echo.Context) error {
	user, err := getUserId(c)
	if !isProjectExists(c.Param("project"), user["_id"].(string)) {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find project",
			Data:    nil,
		})
		return nil
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return err
	}
	database := new(database)
	database.Reads = 0
	database.Creates = 0
	database.Updates = 0
	database.Deletes = 0
	err = c.Bind(database)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	database.Project = c.Param("project")
	if err = c.Validate(database); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
		})
		return nil
	}
	err = applicationDatabase.C("databases").Insert(database)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot insert project to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "project successfully inserted",
		Data:    database,
	})
	return nil
}

func readDatabasesController(c echo.Context) error {
	user, err := getUserId(c)
	if !isProjectExists(c.Param("project"), user["_id"].(string)) {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find project",
			Data:    nil,
		})
		return nil
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	var databases []database
	err = applicationDatabase.C("databases").Find(echo.Map{
		"project": c.Param("project"),
	}).All(&databases)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot query databases",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusOK, response{
		Success: true,
		Message: "Successfully queried databases",
		Data:    databases,
	})
	return nil
}

func updateDatabaseController(c echo.Context) error {
	user, err := getUserId(c)
	if !isProjectExists(c.Param("project"), user["_id"].(string)) {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find doc",
			Data:    nil,
		})
		return nil
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	doc := new(database)
	err = c.Bind(&doc)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	doc.Project = c.Param("project")
	if err = c.Validate(doc); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
		})
		return nil
	}
	var db database
	err = applicationDatabase.C("databases").Find(echo.Map{
		"project": c.Param("project"),
	}).One(&db)
	doc.Reads = db.Reads
	doc.Updates = db.Updates
	doc.Creates = db.Creates
	doc.Deletes = db.Deletes
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	query["project"] = c.Param("project")
	err = applicationDatabase.C("databases").Update(query, doc)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot update doc to doc",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Database successfully updated",
		Data:    doc,
	})
	return nil
}

func deleteDatabaseController(c echo.Context) error {
	user, err := getUserId(c)
	if !isProjectExists(c.Param("project"), user["_id"].(string)) {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find project",
			Data:    nil,
		})
		return nil
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot authorize user",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	query["project"] = c.Param("project")
	err = applicationDatabase.C("databases").Remove(query)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot delete database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Database successfully delete",
		Data:    nil,
	})
	return nil
}
