package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/robertkrimen/otto"
)

func permit(database database, collectionName string, user jwt.MapClaims, action string, document interface{}, id string) bool {
	rule := database.Rules[collectionName+":"+action]
	if rule == nil {
		return false
	}
	vm := otto.New()
	_ = vm.Set("user", user)
	_ = vm.Set("action", action)
	_ = vm.Set("document", document)
	_ = vm.Set("id", id)
	_ = vm.Set("get", func(call otto.FunctionCall) otto.Value {
		collectionName := call.Argument(0).String()
		query, _ := call.Argument(1).Export()
		var all []interface{}
		if database.Url == "" {
			_ = databaseSession.DB(database.Name).C(collectionName).Find(query).All(&all)
		} else {
			session, _ := mgo.Dial(database.Url)
			_ = session.DB("").C(collectionName).Find(query).All(&all)
		}
		value, _ := otto.ToValue(all)
		return value
	})
	result, err := vm.Run(rule)
	if err != nil {
		fmt.Println(err)
		return false
	}
	allowed, err := result.ToBoolean()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return allowed
}
