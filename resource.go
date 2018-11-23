package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"net/http"
)

func createResourceController(c echo.Context) error {
	databaseId := c.Param("database")
	collectionName := c.Param("collection")
	var database database
	//user, _ := getUserId(c)
	if err := applicationDatabase.C("databases").Find(echo.Map{
		"_id": bson.ObjectIdHex(databaseId),
	}).One(&database); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find database",
			Data:    nil,
		})
		return nil
	}
	var resource interface{}
	err := c.Bind(&resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	err = databaseSession.DB(database.Name).C(collectionName).Insert(resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot insert resource to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Resource successfully inserted",
		Data:    resource,
	})
	return nil
}

func readResourcesController(c echo.Context) error {
	databaseId := c.Param("database")
	collectionName := c.Param("collection")
	var database database
	//user, _ := getUserId(c)
	if err := applicationDatabase.C("databases").Find(echo.Map{
		"_id": bson.ObjectIdHex(databaseId),
	}).One(&database); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find database",
			Data:    nil,
		})
		return nil
	}
	iter := databaseSession.DB(database.Name).C(collectionName).Find(echo.Map{}).Iter()
	var resource interface{}
	var resources []interface{}
	for iter.Next(&resource) {
		resources = append(resources, resource)
	}
	_ = c.JSON(http.StatusOK, response{
		Success: true,
		Message: "Successfully queried resources",
		Data:    resources,
	})
	return nil
}

func updateResourceController(c echo.Context) error {
	databaseId := c.Param("database")
	collectionName := c.Param("collection")
	var database database
	//user, _ := getUserId(c)
	if err := applicationDatabase.C("databases").Find(echo.Map{
		"_id": bson.ObjectIdHex(databaseId),
	}).One(&database); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find database",
			Data:    nil,
		})
		return nil
	}
	var resource interface{}
	err := c.Bind(&resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	err = databaseSession.DB(database.Name).C(collectionName).Update(query, resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot update resource to database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Resource successfully updated",
		Data:    resource,
	})
	return nil
}

func deleteResourceController(c echo.Context) error {
	databaseId := c.Param("database")
	collectionName := c.Param("collection")
	var database database
	//user, _ := getUserId(c)
	if err := applicationDatabase.C("databases").Find(echo.Map{
		"_id": bson.ObjectIdHex(databaseId),
	}).One(&database); err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot find database",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	err := databaseSession.DB(database.Name).C(collectionName).Remove(query)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot delete resource from database",
			Data:    nil,
		})
		return nil
	}
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Resource successfully delete",
		Data:    nil,
	})
	return nil
}
