package main

import (
	"crypto/tls"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mlsquires/socketio"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type (
	appValidator struct {
		validator *validator.Validate
	}
)

func (cv *appValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

var rest *echo.Echo

func bootstrapRestServer() {
	rest = echo.New()
	rest.Validator = &appValidator{validator: validator.New()}
	handleRestRoutes()
}

var sockets *socketio.Server

func handleRestRoutes() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	rest.Use(middleware.CORS())

	dashboard, _ := url.Parse("https://enbase-dashboard.netlify.com")
	dashboardProxy := httputil.NewSingleHostReverseProxy(dashboard)
	rest.GET("/*", func(context echo.Context) error {
		context.Request().Host = "enbase-dashboard.netlify.com"
		dashboardProxy.ServeHTTP(context.Response().Writer, context.Request())
		return nil
	})

	rest.POST("/auth/session", loginController)
	rest.POST("/auth/user", registerController)
	rest.POST("/system/projects", createProjectController)
	rest.GET("/system/projects", readProjectsController)
	rest.PUT("/system/projects/:id", updateProjectController)
	rest.DELETE("/system/projects/:id", deleteProjectController)
	rest.POST("/system/projects/:project/databases", createDatabaseController)
	rest.GET("/system/projects/:project/databases", readDatabasesController)
	rest.PUT("/system/projects/:project/databases/:id", updateDatabaseController)
	rest.DELETE("/system/projects/:project/databases/:id", deleteDatabaseController)

	rest.POST("/apps/:database/:collection", createResourceController)
	rest.GET("/apps/:database/:collection", readResourcesController)
	rest.PUT("/apps/:database/:collection/:id", updateResourceController)
	rest.DELETE("/apps/:database/:collection/:id", deleteResourceController)
	rest.GET("/apps/:database/:collection/stream/:action", changesController)
	rest.GET("/ws", crudBusController)
	sockets = configureSockets()
	rest.Any("/socket.io/", func(context echo.Context) error {
		(&corsSocketsServer{
			Server: sockets,
		}).ServeHTTP(context.Response().Writer, context.Request())
		return nil
	})
}

type corsSocketsServer struct {
	Server *socketio.Server
}

func (s *corsSocketsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}
