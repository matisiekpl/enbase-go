package main

import (
	"github.com/globalsign/mgo"
	"os"
)

var databaseSession *mgo.Session
var applicationDatabase *mgo.Database

func connectToDatabase() {
	mongoUrl := os.Getenv("mongo")
	if mongoUrl == "" {
		mongoUrl = "localhost:27017"
	}
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		panic(err)
	}
	databaseSession = session
	applicationDatabase = session.DB("__enbase")
}
