package main

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"github.com/gbrlsnchs/jwt"
	"github.com/graphql-go/graphql"
	"strconv"
	"time"
)

var projectType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Project",
	Fields: graphql.Fields{
		"id":     &graphql.Field{Type: graphql.Int},
		"name":   &graphql.Field{Type: graphql.String},
		"schema": &graphql.Field{Type: graphql.String},
		"mongo":  &graphql.Field{Type: graphql.String},
	},
})

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.Int},
		"name":     &graphql.Field{Type: graphql.String},
		"email":    &graphql.Field{Type: graphql.String},
		"password": &graphql.Field{Type: graphql.String},
		"project":  &graphql.Field{Type: projectType},
		"is_admin": &graphql.Field{Type: graphql.Boolean},
		"projects": &graphql.Field{Type: graphql.NewList(projectType),
			Resolve: func(params graphql.ResolveParams) (i interface{}, e error) {
				var projects []Project
				database.db.Model(&Project{}).Where("user_id = ?", params.Source.(User).ID).Find(&projects)
				return projects, nil
			}},
	},
})

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"me": &graphql.Field{
			Type:        userType,
			Description: "fetch information about currently signed in user",
			Resolve: func(params graphql.ResolveParams) (i interface{}, e error) {
				user, err := Authenticate(params.Context.Value("auth").(string))
				if err != nil {
					return nil, err
				}
				return user, nil
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"login": &graphql.Field{
			Type:        graphql.String,
			Description: "login with email and password",
			Args: graphql.FieldConfigArgument{
				"email":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"password":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"project_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"is_admin":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Boolean)},
			},
			Resolve: LoginResolver,
		},
		"register": &graphql.Field{
			Type:        graphql.String,
			Description: "register with email and password",
			Args: graphql.FieldConfigArgument{
				"name":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"email":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"password":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"project_id": &graphql.ArgumentConfig{Type: graphql.Int},
				"is_admin":   &graphql.ArgumentConfig{Type: graphql.Boolean},
			},
			Resolve: RegisterResolver,
		},
		"createProject": &graphql.Field{
			Type:        projectType,
			Description: "create new project",
			Args: graphql.FieldConfigArgument{
				"name":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"schema": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"mongo":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(params graphql.ResolveParams) (i interface{}, e error) {
				user, err := Authenticate(params.Context.Value("auth").(string))
				if err != nil {
					return nil, err
				}
				project := Project{
					Name:   params.Args["name"].(string),
					Schema: params.Args["schema"].(string),
					Mongo:  params.Args["mongo"].(string),
					UserID: user.ID,
				}
				if ValidateSchema(project.Schema) != nil {
					return nil, ValidateSchema(project.Schema)
				}
				database.db.Create(&project)
				go project.Deploy()
				return project, nil
			},
		},
		"updateProject": &graphql.Field{
			Type:        projectType,
			Description: "update project",
			Args: graphql.FieldConfigArgument{
				"id":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"schema": &graphql.ArgumentConfig{Type: graphql.String},
				"mongo":  &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(params graphql.ResolveParams) (i interface{}, e error) {
				user, err := Authenticate(params.Context.Value("auth").(string))
				if err != nil {
					return nil, err
				}
				var project Project
				var schema string
				var mongo string
				if params.Args["schema"] != nil {
					schema = params.Args["schema"].(string)
				}
				if params.Args["mongo"] != nil {
					mongo = params.Args["mongo"].(string)
				}
				if ValidateSchema(schema) != nil {
					return nil, ValidateSchema(schema)
				}
				database.db.Model(&Project{}).Where("user_id = ? AND id = ?", user.ID, params.Args["id"].(int)).UpdateColumns(&Project{
					Schema: schema,
					Mongo:  mongo,
				}).First(&project)
				go project.Deploy()
				return project, nil
			},
		},
		"deleteProject": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "delete project",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
			},
			Resolve: func(params graphql.ResolveParams) (i interface{}, e error) {
				user, err := Authenticate(params.Context.Value("auth").(string))
				if err != nil {
					return nil, err
				}
				database.db.Model(&Project{}).Where("user_id = ? AND id = ?", user.ID, params.Args["id"].(int)).Delete(&Project{})
				return true, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

func Authenticate(token string) (User, error) {
	hs256 := jwt.NewHMAC(jwt.SHA256, []byte("secret"))
	raw, err := jwt.Parse([]byte(token))
	if err != nil {
		return User{}, errors.New("cannot parse token")
	}
	if err = raw.Verify(hs256); err != nil {
		return User{}, errors.New("cannot verify token")
	}
	var payload jwt.Payload
	if _, err = raw.Decode(&payload); err != nil {
		return User{}, errors.New("cannot decode token")
	}
	var user User
	var count int
	query := database.db.Model(&User{}).Where("id = ?", payload.JWTID)
	query.Count(&count)
	if count < 1 {
		return User{}, errors.New("cannot find given user")
	} else {
		query.First(&user)
		return user, nil
	}
}

func LoginResolver(params graphql.ResolveParams) (i interface{}, e error) {
	hash := sha512.New()
	hash.Write([]byte(params.Args["password"].(string)))
	var user User
	query := database.db.Model(&User{})
	if !params.Args["is_admin"].(bool) {
		query = query.Where("project_id = ?", params.Args["project_id"])
	}
	query = query.Where("email = ? AND password = ?", params.Args["email"], hex.EncodeToString(hash.Sum(nil)))
	var count int
	query.Count(&count)
	if count < 1 {
		return nil, errors.New("cannot find user with given credentials")
	}
	query.First(&user)
	now := time.Now()
	token, err := jwt.Sign(jwt.Header{}, jwt.Payload{
		Issuer:         "enbase",
		Subject:        "users",
		Audience:       jwt.Audience{"https://enteam.me"},
		ExpirationTime: now.Add(24 * 30 * 12 * time.Hour).Unix(),
		NotBefore:      now.Add(30 * time.Minute).Unix(),
		IssuedAt:       now.Unix(),
		JWTID:          strconv.FormatUint(uint64(user.ID), 10),
	}, jwt.NewHMAC(jwt.SHA256, []byte("secret")))
	if err != nil {
		return nil, errors.New("cannot sign token")
	}
	return string(token), nil
}

func RegisterResolver(params graphql.ResolveParams) (i interface{}, e error) {
	query := database.db.Model(&User{})
	if !params.Args["is_admin"].(bool) {
		query = query.Where("project_id = ?", params.Args["project_id"])
	}
	var count int
	query.Where("email = ?", params.Args["email"]).Count(&count)
	if count > 0 {
		return nil, errors.New("user with given email already exists")
	}
	if params.Args["project_id"] == nil {
		params.Args["project_id"] = -1
	}
	if params.Args["is_admin"] == nil {
		params.Args["is_admin"] = false
	}
	hash := sha512.New()
	hash.Write([]byte(params.Args["password"].(string)))
	user := User{
		Name:      params.Args["name"].(string),
		Email:     params.Args["email"].(string),
		ProjectID: params.Args["project_id"].(int),
		IsAdmin:   params.Args["is_admin"].(bool),
		Password:  hex.EncodeToString(hash.Sum(nil)),
		Projects:  []Project{},
	}
	database.db.Create(&user)
	now := time.Now()
	token, err := jwt.Sign(jwt.Header{}, jwt.Payload{
		Issuer:         "enbase",
		Subject:        "users",
		Audience:       jwt.Audience{"https://enteam.me"},
		ExpirationTime: now.Add(24 * 30 * 12 * time.Hour).Unix(),
		NotBefore:      now.Add(30 * time.Minute).Unix(),
		IssuedAt:       now.Unix(),
		JWTID:          strconv.FormatUint(uint64(user.ID), 10),
	}, jwt.NewHMAC(jwt.SHA256, []byte("secret")))
	if err != nil {
		return nil, errors.New("cannot sign token")
	}
	return string(token), nil
}
