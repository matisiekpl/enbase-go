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
	var resource echo.Map
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
	id := bson.NewObjectId()
	resource["_id"] = id
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
		DocumentId:     resource["_id"].(bson.ObjectId).Hex(),
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
		if !permit(database, collectionName, user, "delete", nil, c.Param("id")) || (limited && database.Deletes <= 0) {
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
				query["_id"] = bson.ObjectIdHex(change.DocumentId)
				accessible := false
				if database.Url == "" {
					count, _ := databaseSession.DB(database.Id.Hex()).C(change.CollectionName).Find(&query).Count()
					if count > 0 {
						accessible = true
					}
				} else {
					session, _ := mgo.Dial(database.Url)
					count, _ := session.DB("").C(change.CollectionName).Find(&query).Count()
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

type busRequest struct {
	DatabaseId     string      `json:"database_id"`
	CollectionName string      `json:"collection_name"`
	Action         string      `json:"action"`
	Document       echo.Map    `json:"document"`
	DocumentId     string      `json:"document_id"`
	Token          string      `json:"token"`
	Skip           int         `json:"skip"`
	Limit          int         `json:"limit"`
	RequestId      string      `json:"request_id"`
	MasterKey      string      `json:"master_key"`
	Done           bool        `json:"done"`
	Err            string      `json:"err"`
	Response       interface{} `json:"response"`
}

func crudBusController(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			msg := ""
			err := websocket.Message.Receive(ws, &msg)
			if err == nil && msg != "" {
				var request busRequest
				_ = json.Unmarshal([]byte(msg), &request)
				switch request.Action {
				case "create":
					databaseId := request.DatabaseId
					collectionName := request.CollectionName
					var database database
					user, _ := getUserId(c)
					if err := applicationDatabase.C("databases").Find(echo.Map{
						"_id": bson.ObjectIdHex(databaseId),
					}).One(&database); err != nil {
						request.Done = false
						request.Err = "Cannot find database"
					}
					if request.MasterKey != database.MasterKey {
						if !permit(database, collectionName, user, "create", request.Document, "") || (limited && database.Creates <= 0) {
							request.Done = false
							request.Err = "Access denied"
						}
					}
					id := bson.NewObjectId()
					request.Document["_id"] = id
					if database.Url == "" {
						err = databaseSession.DB(database.Id.Hex()).C(collectionName).Insert(request.Document)
					} else {
						session, _ := mgo.Dial(database.Url)
						err = session.DB("").C(collectionName).Insert(request.Document)
					}
					if err != nil {
						request.Done = false
						request.Err = "Cannot insert resource to database"
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
						Document:       request.Document,
						DocumentId:     request.Document["_id"].(bson.ObjectId).Hex(),
						Action:         "create",
						DatabaseId:     databaseId,
					})
					request.Done = true
					request.Err = "Resource successfully inserted"
					requestResponse, _ := json.Marshal(request)
					_ = websocket.Message.Send(ws, string(requestResponse))
				case "update":
					databaseId := request.DatabaseId
					collectionName := request.CollectionName
					var database database
					user, _ := getUserId(c)
					if err := applicationDatabase.C("databases").Find(echo.Map{
						"_id": bson.ObjectIdHex(databaseId),
					}).One(&database); err != nil {
						request.Done = false
						request.Err = "Cannot find database"
					}
					if request.MasterKey != database.MasterKey {
						if !permit(database, collectionName, user, "update", request.Document, "") || (limited && database.Updates <= 0) {
							request.Done = false
							request.Err = "Access denied"
						}
					}
					query := echo.Map{}
					query["_id"] = bson.ObjectIdHex(request.DocumentId)
					if database.Url == "" {
						err = databaseSession.DB(database.Id.Hex()).C(collectionName).Update(query, request.Document)
					} else {
						session, _ := mgo.Dial(database.Url)
						err = session.DB("").C(collectionName).Update(query, request.Document)
					}
					if err != nil {
						request.Done = false
						request.Err = "Cannot update resource"
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
						DocumentId:     request.DocumentId,
						Document:       request.Document,
						Action:         "update",
						DatabaseId:     databaseId,
					})
					request.Done = true
					request.Err = "Resource successfully updated"
					requestResponse, _ := json.Marshal(request)
					_ = websocket.Message.Send(ws, string(requestResponse))
				case "delete":
					databaseId := request.DatabaseId
					collectionName := request.CollectionName
					var database database
					user, _ := getUserId(c)
					if err := applicationDatabase.C("databases").Find(echo.Map{
						"_id": bson.ObjectIdHex(databaseId),
					}).One(&database); err != nil {
						request.Done = false
						request.Err = "Cannot find database"
					}
					if request.MasterKey != database.MasterKey {
						if !permit(database, collectionName, user, "delete", request.Document, request.DocumentId) || (limited && database.Deletes <= 0) {
							request.Done = false
							request.Err = "Access denied"
						}
					}
					query := echo.Map{}
					query["_id"] = bson.ObjectIdHex(request.DocumentId)
					var err error
					if database.Url == "" {
						err = databaseSession.DB(database.Id.Hex()).C(collectionName).Remove(query)
					} else {
						session, _ := mgo.Dial(database.Url)
						err = session.DB("").C(collectionName).Remove(query)
					}
					if err != nil {
						request.Done = false
						request.Err = "Cannot delete resource"
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
						DocumentId:     request.DocumentId,
						Action:         "delete",
						DatabaseId:     databaseId,
					})
					request.Done = true
					request.Err = "Resource successfully deleted"
					requestResponse, _ := json.Marshal(request)
					_ = websocket.Message.Send(ws, string(requestResponse))
				case "read":
					databaseId := request.DatabaseId
					collectionName := request.CollectionName
					var database database
					user, _ := getUserId(c)
					if err := applicationDatabase.C("databases").Find(echo.Map{
						"_id": bson.ObjectIdHex(databaseId),
					}).One(&database); err != nil {
						request.Done = false
						request.Err = "Cannot find database"
					}
					var iter *mgo.Iter
					if database.Url == "" {
						iter = databaseSession.DB(database.Id.Hex()).C(collectionName).Find(request.Document).Iter()
					} else {
						session, _ := mgo.Dial(database.Url)
						iter = session.DB("").C(collectionName).Find(request.Document).Iter()
					}
					var resource interface{}
					var resources []interface{}
					resourcesLimit := 50
					resourcesSkip := 0
					if request.Limit != 0 {
						resourcesLimit = request.Limit
					}
					if request.Skip != 0 {
						resourcesSkip = request.Skip
					}
					resourcesCount := 0
					for resourcesCount < resourcesLimit && iter.Next(&resource) {
						if request.MasterKey == database.MasterKey || (permit(database, collectionName, user, "read", resource, "") && !(limited && database.Reads <= 0)) {
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
					request.Response = resources
					request.Done = true
					request.Err = "Resources successfully queried"
					requestResponse, _ := json.Marshal(request)
					_ = websocket.Message.Send(ws, string(requestResponse))
				}
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
