package main

import "github.com/globalsign/mgo"

var databaseSession *mgo.Session
var applicationDatabase *mgo.Database

func connectToDatabase() {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	databaseSession = session
	applicationDatabase = session.DB("__enbase")
}
