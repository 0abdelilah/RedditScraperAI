package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"redditor/handlers"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	handlers.Client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	// Ping MongoDB to verify connection
	if err := handlers.Client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}

	fmt.Println("Connected to MongoDB")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /static/", handlers.StaticHandler)
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/analyse", handlers.AnalyseHandler)
	mux.HandleFunc("/gethistory", handlers.GetHistoryHandler)
	mux.HandleFunc("/history", handlers.HistoryHandler)

	fmt.Println("Started: http://localhost:8080/")
	http.ListenAndServe(":8080", mux)
}
