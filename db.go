package main

import (
	"github.com/globalsign/mgo"
	"os"
	"strconv"
)

var databaseSession *mgo.Session
var applicationDatabase *mgo.Database
var limited bool

func connectToDatabase() {
	limited, _ = strconv.ParseBool(os.Getenv("LIMITED"))
	mongoUrl := os.Getenv("MONGO")
	mongoName := os.Getenv("MONGO_NAME")
	if mongoUrl == "" {
		mongoUrl = "localhost:27017"
	}
	if mongoName == "" {
		mongoName = "__enbase"
	}
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		panic(err)
	}
	databaseSession = session
	applicationDatabase = session.DB(mongoName)
}
