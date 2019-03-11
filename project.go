package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
)

type Project struct {
	Id          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Author      string        `json:"author"`
	Rules       echo.Map      `json:"rules"`
	Url         string        `json:"url"`
}
