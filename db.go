package main

import (
	"github.com/globalsign/mgo"
	"os"
)

var databaseSession *mgo.Session
var applicationDatabase *mgo.Database

func connectToDatabase() {
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
