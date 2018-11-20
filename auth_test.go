package main

import (
	"github.com/gavv/httpexpect"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ClearDatabase() {
	_ = applicationDatabase.DropDatabase()
}

func TestSignInWithIncorrectCredentials(t *testing.T) {
	ConnectToDatabase()
	ClearDatabase()
	BootstrapRestServer()
	server := httptest.NewServer(rest)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	e.POST("/auth/session").WithJSON(User{
		Email:    "test@test.pl",
		Password: "1234",
	}).Expect().Status(http.StatusBadRequest)
}

func TestSignInWithIncorrectBody(t *testing.T) {
	ConnectToDatabase()
	ClearDatabase()
	BootstrapRestServer()
	server := httptest.NewServer(rest)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	e.POST("/auth/session").WithText("test").Expect().Status(http.StatusBadRequest)
}

func TestSignUpWithIncorrectBody(t *testing.T) {
	ConnectToDatabase()
	ClearDatabase()
	BootstrapRestServer()
	server := httptest.NewServer(rest)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	e.POST("/auth/user").WithText("test").Expect().Status(http.StatusBadRequest)
}

func TestSignInWithCorrectCredentials(t *testing.T) {
	ConnectToDatabase()
	ClearDatabase()
	BootstrapRestServer()
	server := httptest.NewServer(rest)
	defer server.Close()
	e := httpexpect.New(t, server.URL)
	e.POST("/auth/user").WithJSON(User{
		Email:    "test@test.pl",
		Password: "1234",
		Role:     "admin",
	}).Expect().Status(http.StatusOK)
	e.POST("/auth/session").WithJSON(User{
		Email:    "test@test.pl",
		Password: "1234",
		Role:     "",
	}).Expect().Status(http.StatusOK).Body().Contains("token")
}
