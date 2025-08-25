package scrapers

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

func GetSubreddits(keyword string, limit string) ([]string, error) {
	url := "https://www.reddit.com/subreddits/search.json?q=" + url.QueryEscape(keyword) + "&limit=" + limit

	b, err := FetchRedditJSON(url, 2)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	children := gjson.GetBytes(b, "data.children").Array()
	subredditNames := []string{}
	for _, c := range children {
		if c.Get("data.subreddit_type").String() == "public" {
			subredditNames = append(subredditNames, c.Get("data.display_name").String())
		}
	}

	return subredditNames, nil
}

func GetPostsIDs(subreddit string, limit string) ([]string, error) {
	url := "https://www.reddit.com/r/" + subreddit + "/hot.json?limit=" + limit

	b, err := FetchRedditJSON(url, 1)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	results := gjson.GetBytes(b, "data.children.#.data.id").Array()

	// Convert gjson.Result to []string
	subredditNames := make([]string, len(results))
	for i, r := range results {
		subredditNames[i] = r.String()
	}

	return subredditNames, nil
}

type PostData struct {
	Subreddit  string `json:"subreddit"`
	PostID     string `json:"post_id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Comments   string `json:"comments"`
	CommentsSl []string
}

func GetPostInfo(subreddit, postID string) (*PostData, error) {
	url := "https://www.reddit.com/r/" + subreddit + "/comments/" + postID + ".json"

	b, err := FetchRedditJSON(url, 1)
	if err != nil {
		return nil, err
	}

	title := gjson.GetBytes(b, "0.data.children.0.data.title").String()
	body := gjson.GetBytes(b, "0.data.children.0.data.selftext").String()

	var extractReplies func(r gjson.Result) []string
	extractReplies = func(r gjson.Result) []string {
		comments := []string{}
		r.Get("data.children").ForEach(func(_, c gjson.Result) bool {
			text := html.UnescapeString(c.Get("data.body").String())
			if text != "" {
				comments = append(comments, text)
			}
			if replies := c.Get("data.replies"); replies.Exists() && replies.Type == gjson.JSON {
				comments = append(comments, extractReplies(replies)...)
			}
			return true
		})
		return comments
	}

	commentsSlice := extractReplies(gjson.GetBytes(b, "1"))
	commentsJoined := strings.Join(commentsSlice, "\n")

	return &PostData{
		Subreddit:  subreddit,
		PostID:     postID,
		Title:      html.UnescapeString(title),
		Body:       html.UnescapeString(body),
		Comments:   commentsJoined,
		CommentsSl: commentsSlice,
	}, nil
}

type PostResult struct {
	Link              string
	Community         string  `json:"Community"`
	Summary           string  `json:"Summary"`
	MicrosaasSolution string  `json:"MicrosaasSolution"`
	Rating            float64 `json:"Rating"`
}

type KeywordResult struct {
	Keyword string       `json:"keyword"`
	Posts   []PostResult `json:"posts"`
}

type Prompt struct {
	Id       string
	Title    string
	Body     string
	Comments []string
}

func ScrapeReddit(ctx context.Context, keyword string, ch chan<- KeywordResult, maxPosts int) error {
	// Get public subreddits matching keyword
	subreddits, err := GetSubreddits(keyword, "3")
	if err != nil {
		fmt.Println("Error fetching subreddits:", err)
		return err
	}
	if len(subreddits) == 0 {
		fmt.Println("No subreddits found. Exiting.")
		return nil
	}

	fmt.Printf("Found %d public subreddits\n", len(subreddits))
	fmt.Println(subreddits)

	count := 0

	// Loop over all subreddits sequentially
	for _, subreddit := range subreddits {
		select {
		case <-ctx.Done():
			fmt.Println("Scraping cancelled")
			return nil
		default:
		}

		postIDs, err := GetPostsIDs(subreddit, "10")
		if err != nil {
			fmt.Println("Error fetching posts for subreddit", subreddit, ":", err)
			continue
		}
		if len(postIDs) == 0 {
			fmt.Println("No posts found in subreddit", subreddit)
			continue
		}

		fmt.Printf("Found %d post_ids for %s\n", len(postIDs), subreddit)

		// Loop through posts sequentially
		for _, postID := range postIDs {
			select {
			case <-ctx.Done():
				fmt.Println("Scraping cancelled")
				return nil
			default:
			}

			RawPost, err := GetPostInfo(subreddit, postID)
			if err != nil {
				fmt.Println("Error fetching post", postID, ":", err)
				continue
			}

			textToAnalyze := RawPost.Title + "\n\n" + RawPost.Body + "\n\n" + RawPost.Comments

			analysis := analyzePostContent(textToAnalyze)
			if analysis.Has_pain_point {
				link := "https://www.reddit.com/r/" + subreddit + "/comments/" + postID + ""

				post := PostResult{
					Link:              link,
					Community:         RawPost.Subreddit,
					Summary:           analysis.Summary,
					MicrosaasSolution: analysis.Microsaas_solution,
					Rating:            analysis.Rating,
				}

				keywordOutput := KeywordResult{
					Keyword: keyword,
					Posts:   []PostResult{post},
				}

				ch <- keywordOutput
				count++
			}

			if count >= maxPosts {
				fmt.Println("Reached maxPosts limit, stopping scraping")
				return nil
			}
		}
	}
	return nil
}
