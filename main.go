package main

import (
	"log"
	"net/http"
	"os"
	"posts/graph"
	"posts/storage"
	"posts/storage/inmemory"
	"posts/storage/postgres"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	serverPort  string = "8080"
	storageType string = "IN_MEMORY"
	dbHost      string = "postgres"
	dbUser      string = "user"
	dbPassword  string = "password"
	dbName      string = "posts_db"
	dbPort      string = "5432"
)

func loadEnv() {
	if port := os.Getenv("PORT"); port != "" {
		serverPort = port
	}
	if storage := os.Getenv("STORAGE_TYPE"); storage != "" {
		storageType = storage
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		dbHost = host
	}
	if user := os.Getenv("DB_USER"); user != "" {
		dbUser = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		dbPassword = password
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		dbName = db
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		dbPort = port
	}
}

func makeStorage() (storage.Storage, error) {
	switch storageType {
	case "POSTGRES":
		return postgres.NewPostgreStorage(dbHost, dbUser, dbPassword, dbName, dbPort)
	case "IN_MEMORY":
		return inmemory.NewInMemoryStorage()
	default:
		panic("invalid storage type")
	}
}

func main() {
	loadEnv()

	postStorage, err := makeStorage()
	if err != nil {
		log.Fatalf("error creating storage: %v", err)
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(postStorage)}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	srv.Use(extension.FixedComplexityLimit(16))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Server started on http://localhost:%s/", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}
