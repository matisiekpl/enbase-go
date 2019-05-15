package main

import "github.com/globalsign/mgo"

var Database *mgo.Database

func ConnectToDatabase() {
	session, err := mgo.Dial("localhost/enbase-ese-management")
	if err != nil {
		panic(err)
	}
	Database = session.DB("")
}
