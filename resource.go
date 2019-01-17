package main

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/joncalhoun/qson"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
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
	if c.Request().Header.Get("X-master-key") != database.MasterKey {
		if !permit(database, collectionName, user, "create", resource, "") || (limited && database.Creates <= 0) {
			_ = c.JSON(http.StatusBadRequest, response{
				Success: false,
				Message: "Access denied",
				Data:    nil,
			})
			return nil
		}
	}
	if database.Url == "" {
		err = databaseSession.DB(database.Id.Hex()).C(collectionName).Insert(resource)
	} else {
		session, _ := mgo.Dial(database.Url)
		err = session.DB("").C(collectionName).Insert(resource)
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot insert resource to database",
			Data:    nil,
		})
		return nil
	}
	if limited {
		database.Creates--
	}
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
	var iter *mgo.Iter
	if database.Url == "" {
		iter = databaseSession.DB(database.Id.Hex()).C(collectionName).Find(query).Iter()
	} else {
		session, _ := mgo.Dial(database.Url)
		iter = session.DB("").C(collectionName).Find(query).Iter()
	}
	var resource interface{}
	var resources []interface{}
	resourcesLimit := 50
	resourcesSkip := 0
	if c.Request().Header.Get("X-enbase-limit") != "" {
		resourcesLimit, _ = strconv.Atoi(c.Request().Header.Get("X-enbase-limit"))
	}
	if c.Request().Header.Get("X-enbase-skip") != "" {
		resourcesSkip, _ = strconv.Atoi(c.Request().Header.Get("X-enbase-skip"))
	}
	resourcesCount := 0
	for resourcesCount < resourcesLimit && iter.Next(&resource) {
		if c.Request().Header.Get("X-master-key") == database.MasterKey || (permit(database, collectionName, user, "read", resource, "") && !(limited && database.Reads <= 0)) {
			resourcesSkip--
			if resourcesSkip < 0 {
				if limited {
					database.Reads--
				}
				query := echo.Map{}
				query["_id"] = database.Id
				_ = applicationDatabase.C("databases").Update(query, database)
				resources = append(resources, resource)
				resourcesCount++
			}
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
	if c.Request().Header.Get("X-master-key") != database.MasterKey {
		if !permit(database, collectionName, user, "update", resource, "") || (limited && database.Updates <= 0) {
			_ = c.JSON(http.StatusBadRequest, response{
				Success: false,
				Message: "Access denied",
				Data:    nil,
			})
			return nil
		}
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	if database.Url == "" {
		err = databaseSession.DB(database.Id.Hex()).C(collectionName).Update(query, resource)
	} else {
		session, _ := mgo.Dial(database.Url)
		err = session.DB("").C(collectionName).Update(query, resource)
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot update resource to database",
			Data:    nil,
		})
		return nil
	}
	if limited {
		database.Updates--
	}
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
	if c.Request().Header.Get("X-master-key") != database.MasterKey {
		if !permit(database, collectionName, user, "delete", nil, "") || (limited && database.Deletes <= 0) {
			_ = c.JSON(http.StatusBadRequest, response{
				Success: false,
				Message: "Access denied",
				Data:    nil,
			})
			return nil
		}
	}
	query := echo.Map{}
	query["_id"] = bson.ObjectIdHex(c.Param("id"))
	var err error
	if database.Url == "" {
		err = databaseSession.DB(database.Id.Hex()).C(collectionName).Remove(query)
	} else {
		session, _ := mgo.Dial(database.Url)
		err = session.DB("").C(collectionName).Remove(query)
	}
	if err != nil {
		_ = c.JSON(http.StatusBadRequest, response{
			Success: false,
			Message: "Cannot delete resource from database",
			Data:    nil,
		})
		return nil
	}
	if limited {
		database.Deletes--
	}
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
				queryJson, _ := qson.ToJSON(c.QueryString())
				var query echo.Map
				_ = json.Unmarshal(queryJson, &query)
				if query == nil {
					query = echo.Map{}
				}
				query["_id"] = change.DocumentId
				accessible := false
				if database.Url == "" {
					count, _ := databaseSession.DB(database.Id.Hex()).C(change.CollectionName).Find(query).Count()
					if count > 0 {
						accessible = true
					}
				} else {
					session, _ := mgo.Dial(database.Url)
					count, _ := session.DB("").C(change.CollectionName).Find(query).Count()
					if count > 0 {
						accessible = true
					}
				}
				if accessible {
					if c.Request().Header.Get("X-master-key") != database.MasterKey {
						if permit(database, c.Param("collection"), nil, "stream", change.Document, change.DocumentId) {
							payload, _ := json.Marshal(change)
							_ = websocket.Message.Send(ws, []byte(payload))
						}
					} else {
						payload, _ := json.Marshal(change)
						_ = websocket.Message.Send(ws, []byte(payload))
					}
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
