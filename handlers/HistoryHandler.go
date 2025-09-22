package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := Client.Database("reddit").Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Read all documents
	var posts []ResultPosts
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, `{"success": false, "error": "Failed to read from DB"}`, http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &posts); err != nil {
		http.Error(w, `{"success": false, "error": "Failed to decode DB results"}`, http.StatusInternalServerError)
		return
	}

	// Return results
	resp := map[string]interface{}{
		"success": true,
		"data":    posts,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("Failed to encode JSON:", err)
	}
}

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	tmpt, err := template.ParseFiles("./templates/history.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	tmpt.Execute(w, nil)
}
