package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"redditor/ai"
	"redditor/scrapers"
	"strings"

	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
)

var Client *mongo.Client

type AIResponseFiltered struct {
	Link           string  `json:"link"`
	Pain_point     string  `json:"pain_point"`
	Classification string  `json:"classification"`
	Problem_type   string  `json:"problem_type"`
	Reoccurrence   float64 `json:"reoccurrence"`
}

type ResultPosts struct {
	Keyword string               `json:"keyword"`
	Posts   []AIResponseFiltered `json:"posts"`
}

func AnalyseHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	m, err := strconv.Atoi(r.URL.Query().Get("max"))
	
	if keyword == "" {
		http.Error(w, "Missing keyword", http.StatusBadRequest)
		return
	}

	if err == nil && m > 0 && m < 10 {
		m = 10
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	posts, err := scrapers.ScrapeReddit(ctx, keyword, m)
	if err != nil {
		fmt.Println("Scrape error:", err)
	}

	postsJSON, _ := json.Marshal(posts)

	result := ai.AnalyzePosts(string(postsJSON))
	var filtered []AIResponseFiltered

	// filter - empty ones, turn id to link
	for i := range result {
		if result[i].Pain_point == "" {
			continue
		}
		sub_id := strings.Split(result[i].Id, "/")
		if len(sub_id) != 2 {
			continue
		}

		link := "https://www.reddit.com/r/" + sub_id[0] + "/comments/" + sub_id[1]

		filtered = append(filtered, AIResponseFiltered{
			Link:           link,
			Pain_point:     result[i].Pain_point,
			Classification: result[i].Classification,
			Problem_type:   result[i].Problem_type,
			Reoccurrence:   result[i].Reoccurrence,
		})
	}

	wrapped := ResultPosts{
		Keyword: keyword,
		Posts:   filtered,
	}

	jsonData, _ := json.Marshal(wrapped)
	os.WriteFile("person.json", jsonData, 0644)

	collection := Client.Database("reddit").Collection("posts")

	// Insert the struct, not JSON bytes
	if _, err := collection.InsertOne(ctx, wrapped); err != nil {
		fmt.Println("Mongo insert error:", err)
	} else {
		fmt.Println("Inserted successfully")
	}

	fmt.Println("Scraping finished or stopped.")
}
