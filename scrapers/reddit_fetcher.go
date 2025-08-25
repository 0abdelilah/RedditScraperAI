package scrapers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const rateLimitFile = "last_request.txt"
const minInterval = 50 * time.Millisecond // 20 req/sec => 50ms per request

// Save current time to file
func saveLastRequestTime(t time.Time) error {
	return os.WriteFile(rateLimitFile, []byte(strconv.FormatInt(t.UnixNano(), 10)), 0644)
}

// Load last request time from file
func loadLastRequestTime() (time.Time, error) {
	data, err := os.ReadFile(rateLimitFile)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil // file doesn't exist yet
		}
		return time.Time{}, err
	}
	ns, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, ns), nil
}

// Wait to respect rate limit
func waitRateLimit() {
	lastTime, err := loadLastRequestTime()
	if err != nil {
		fmt.Println("Warning: cannot read last request time:", err)
		return
	}

	if !lastTime.IsZero() {
		elapsed := time.Since(lastTime)
		if elapsed < minInterval {
			time.Sleep(minInterval - elapsed)
		}
	}
}

var totalRequests int

func incrementTotalRequests() {
	totalRequests++
	fmt.Printf("Total requests made: %d\n", totalRequests)
}

func FetchRedditJSON(endpoint string, maxRetries int) ([]byte, error) {
	client := &http.Client{}
	var resp *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		waitRateLimit()
		incrementTotalRequests() // count request for program run

		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")

		resp, err = client.Do(req)
		saveLastRequestTime(time.Now())

		if err == nil && resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed reading response body: %w", err)
			}
			return body, nil
		}

		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(2 * time.Second) // retry delay
	}

	return nil, fmt.Errorf("failed to fetch URL after %d retries: %v", maxRetries, err)
}
