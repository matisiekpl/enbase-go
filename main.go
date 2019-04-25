package main

import (
	"github.com/globalsign/mgo"
	"github.com/graphql-go/graphql/language/ast"
)

var database Database

func main() {
	database = Database{}
	database.Connect()
	hosting := Hosting{
		Database: database,
	}
	go Init()
	hosting.Listen()
}

func Init() {
	connections = map[uint]*mgo.Database{}
	directives = map[uint]map[string][]*ast.Directive{}
	fieldDirectives = map[uint]map[string]map[string][]*ast.Directive{}
	var projects []Project
	database.db.Find(&projects)
	for _, project := range projects {
		project.Deploy()
	}
}
