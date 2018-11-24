package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/robertkrimen/otto"
)

func permit(database database, collectionName string, user jwt.MapClaims, action string, document interface{}, id string) bool {
	rule := database.Rules[collectionName+":"+action]
	if rule == nil {
		return false
	}
	vm := otto.New()
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
