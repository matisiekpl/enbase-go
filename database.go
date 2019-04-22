package main

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
	"time"
)

type Project struct {
	ID          uint                      `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	DeletedAt   *time.Time                `sql:"index" json:"deleted_at"`
	Name        string                    `json:"name"`
	Schema      string                    `json:"schema" sql:"size:999999"`
	Mongo       string                    `json:"mongo"`
	UserID      uint                      `json:"user_id"`
	Types       map[string]graphql.Output `gorm:"-"`
	Definitions *ast.Document             `gorm:"-"`
}

type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	ProjectID int        `json:"project_id"`
	IsAdmin   bool       `json:"is_admin"`
	Projects  []Project  `json:"projects"`
}

type Database struct {
	db *gorm.DB
}

func (database *Database) Connect() {
	db, err := gorm.Open(os.Getenv("CONTROLLER_DATABASE_DIALECT"), os.Getenv("CONTROLLER_DATABASE_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	database.db = db
	database.db.AutoMigrate(&User{})
	database.db.AutoMigrate(&Project{})
}
