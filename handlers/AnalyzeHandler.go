package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"redditor/scrapers"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
)

var Client *mongo.Client

func AnalyseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		http.Error(w, "Missing keyword", http.StatusBadRequest)
		return
	}

	maxPosts := 10
	if m, err := strconv.Atoi(r.URL.Query().Get("max")); err == nil && m > 0 {
		if m < maxPosts {
			maxPosts = m
		}
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ch := make(chan scrapers.KeywordResult)

	go func() {
		if err := scrapers.ScrapeReddit(ctx, keyword, ch, maxPosts); err != nil {
			fmt.Println("Scrape error:", err)
		}
		close(ch)
	}()

	collection := Client.Database("reddit").Collection("posts")

	for result := range ch {
		jsonData, _ := json.Marshal(result)

		// Stream to frontend
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Insert into MongoDB
		if _, err := collection.InsertOne(ctx, result); err != nil {
			fmt.Println("Mongo insert error:", err)
		}
	}

	// Send "done" event to tell frontend to stop loader
	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	fmt.Println("Scraping finished or stopped.")
}
