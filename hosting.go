package main

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"net/http"
	"strings"
)

type Hosting struct {
	ProviderSchema graphql.Schema
	Database       Database
}

func (hosting *Hosting) Listen() {
	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(context.Background(), "auth", strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", ""))
		h.ContextHandler(ctx, w, r)
	})
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		panic(err)
	}
}
