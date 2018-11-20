package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"strings"
	"time"
)

type LoginResponseData struct {
	Token string `json:"token"`
}

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

func LoginController(httpContext echo.Context) error {
	credentials := echo.Map{}
	err := httpContext.Bind(&credentials)
	if err != nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return err
	}
	var user echo.Map
	email := credentials["email"]
	password := credentials["password"].(string)
	passwordHash := sha512.New()
	_, _ = io.WriteString(passwordHash, password)
	hashedPassword := base64.URLEncoding.EncodeToString(passwordHash.Sum(nil)[:])
	err = applicationDatabase.C("__users").Find(echo.Map{
		"email":    email,
		"password": hashedPassword,
	}).One(&user)
	if err != nil || user == nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot find user with given credentials",
			Data:    nil,
		})
		return err
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["firstName"] = user["firstName"]
	claims["lastName"] = user["lastName"]
	claims["_id"] = user["_id"]
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	tokenString, err := token.SignedString([]byte("jwt-token-secret"))
	if err != nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Error while signing jwt token",
			Data:    nil,
		})
		return err
	}
	data := LoginResponseData{
		Token: tokenString,
	}
	_ = httpContext.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Successfully signed in",
		Data:    data,
	})
	return nil
}

func RegisterController(httpContext echo.Context) error {
	user := new(User)
	err := httpContext.Bind(&user)
	if err != nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot decode body",
			Data:    nil,
		})
		return nil
	}
	if err = httpContext.Validate(user); err != nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Validation failed",
			Data:    nil,
		})
		fmt.Println(err)
		return nil
	}
	email := user.Email
	query := make(echo.Map)
	query["email"] = email
	count, err := applicationDatabase.C("__users").Find(query).Count()
	if err != nil {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Cannot query database",
			Data:    nil,
		})
		return err
	}
	if count == 0 {
		password := user.Password
		passwordHash := sha512.New()
		_, _ = io.WriteString(passwordHash, password)
		user.Password = base64.URLEncoding.EncodeToString(passwordHash.Sum(nil)[:])
		_ = applicationDatabase.C("__users").Insert(user)
		user.Password = ""
		_ = httpContext.JSON(http.StatusOK, Response{
			Success: true,
			Message: "Successfully signed up",
			Data:    user,
		})
		return nil
	} else {
		_ = httpContext.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "User with given email exists",
			Data:    nil,
		})
		return err
	}
}

func GetUserId(c echo.Context) (jwt.MapClaims, error) {
	tokenStr := strings.Replace(c.Request().Header.Get("Authorization"), "Bearer ", "", 1)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte("jwt-token-secret"), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, nil
	}
}
