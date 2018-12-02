package main

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/joncalhoun/qson"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"net/http"
)

type resourceChange struct {
	DatabaseId     string      `json:"database_id"`
	DatabaseName   string      `json:"database_name"`
	CollectionName string      `json:"collection_name"`
	Document       interface{} `json:"document"`
	DocumentId     string      `json:"document_id"`
	Action         string      `json:"action"`
}

func createResourceController(c echo.Context) error {
	databaseId := c.Param("database")
	collectionName := c.Param("collection")
	var database database
	user, _ := getUserId(c)
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
	if !permit(database, collectionName, user, "create", resource, "") || database.Creates <= 0 {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Access denied",
			Data:    nil,
		})
		return nil
	}
	err = databaseSession.DB(database.Id.Hex()).C(collectionName).Insert(resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot insert resource to database",
			Data:    nil,
		})
		return nil
	}
	database.Creates--
	query := echo.Map{}
	query["_id"] = database.Id
	_ = applicationDatabase.C("databases").Update(query, database)
	_ = publishChange(resourceChange{
		DatabaseName:   database.Name,
		CollectionName: collectionName,
		Document:       resource,
		DocumentId:     "",
		Action:         "create",
		DatabaseId:     databaseId,
	})
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
	user, _ := getUserId(c)
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
	queryJson, _ := qson.ToJSON(c.QueryString())
	var query interface{}
	_ = json.Unmarshal(queryJson, &query)
	iter := databaseSession.DB(database.Id.Hex()).C(collectionName).Find(query).Iter()
	var resource interface{}
	var resources []interface{}
	for iter.Next(&resource) {
		if permit(database, collectionName, user, "read", resource, "") || database.Reads <= 0 {
			database.Reads--
			query := echo.Map{}
			query["_id"] = database.Id
			_ = applicationDatabase.C("databases").Update(query, database)
			resources = append(resources, resource)
		}
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
	user, _ := getUserId(c)
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
	if !permit(database, collectionName, user, "update", resource, "") || database.Updates <= 0 {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Access denied",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	err = databaseSession.DB(database.Id.Hex()).C(collectionName).Update(query, resource)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot update resource to database",
			Data:    nil,
		})
		return nil
	}
	database.Updates--
	query = echo.Map{}
	query["_id"] = database.Id
	_ = applicationDatabase.C("databases").Update(query, database)
	_ = publishChange(resourceChange{
		DatabaseName:   database.Name,
		CollectionName: collectionName,
		DocumentId:     c.Param("id"),
		Document:       resource,
		Action:         "update",
		DatabaseId:     databaseId,
	})
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
	user, _ := getUserId(c)
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
	if !permit(database, collectionName, user, "delete", nil, "") || database.Deletes <= 0 {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Access denied",
			Data:    nil,
		})
		return nil
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	err := databaseSession.DB(database.Id.Hex()).C(collectionName).Remove(query)
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot delete resource from database",
			Data:    nil,
		})
		return nil
	}
	database.Deletes--
	query = echo.Map{}
	query["_id"] = database.Id
	_ = applicationDatabase.C("databases").Update(query, database)
	_ = publishChange(resourceChange{
		DatabaseName:   database.Name,
		CollectionName: collectionName,
		Document:       nil,
		DocumentId:     c.Param("id"),
		Action:         "delete",
		DatabaseId:     databaseId,
	})
	_ = c.JSON(http.StatusCreated, response{
		Success: true,
		Message: "Resource successfully delete",
		Data:    nil,
	})
	return nil
}

func changesController(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		sub := localPubsub.Sub("changes")
		for msg := range sub {
			change := msg.(resourceChange)
			if change.DatabaseId == c.Param("database") && change.CollectionName == c.Param("collection") && change.Action == c.Param("action") {
				var database database
				_ = applicationDatabase.C("databases").FindId(bson.ObjectIdHex(c.Param("database"))).One(&database)
				if permit(database, c.Param("collection"), nil, change.Action, change.Document, change.DocumentId) {
					payload, _ := json.Marshal(change)
					_ = websocket.Message.Send(ws, []byte(payload))
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}