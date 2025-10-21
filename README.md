# Redditor — Community Pain Points Analyzer

A small Go web application that scrapes Reddit for posts matching a keyword, sends post data to an AI model for extraction of "pain points", stores results in MongoDB, and exposes a simple web UI to view insights and history.

This README explains how the project works, how to run it locally on Windows (PowerShell), and important notes for development and security.

## What it does

- Searches Reddit for subreddits matching a keyword.
- Fetches posts and comments from those subreddits.
- Calls an AI model to extract concise "pain points", classifications, and problem-type labels from posts.
- Saves results to a MongoDB collection and to `person.json` on each analysis.
- Serves a small frontend to run scrapes and view historical results/analytics.

## Project layout

- `server.go` — application entrypoint and HTTP routing.
- `handlers/` — HTTP handlers for home, analysis, and history endpoints.
- `ai/ai.go` — code that calls Google GenAI to analyze posts.
- `scrapers/` — Reddit scraping utilities and a rate-limited fetcher.
- `templates/` — HTML templates and static assets (CSS/JS).
- `last_request.txt` — persisted timestamp used for simple rate-limiting between runs.

## Requirements

- Go 1.20+ (module requires go 1.23 in go.mod; any recent Go 1.20+ should work).
- A running MongoDB instance accessible at `mongodb://localhost:27017` (or change code in `server.go`).
- Internet access for Reddit scraping and the AI API.

Optional developer tools:

- `curl` or a browser to test endpoints.

## Important security note

The AI client in `ai/ai.go` currently contains an API key string hard-coded in the source. Do NOT commit or expose real API keys. Replace that value with reading from an environment variable or a secure secret manager before using in production.

Search for this line in `ai/ai.go`:

```go
APIKey: "AIzaSyAvqlxoIQqupWEKO_Hd3OgO-pIdUiljnD0",
```

Replace it with something like:

```go
APIKey: os.Getenv("GENAI_API_KEY"),
```

and set `GENAI_API_KEY` in your environment before running.

## Build & run (Windows PowerShell)

1. Open PowerShell and change to the project folder:

```powershell
cd 'C:\Users\gls\Desktop\redditor'
```

2. Ensure MongoDB is running locally on the default port 27017. If not, start or configure it appropriately.

3. (Optional) Set API key as environment variable (recommended):

```powershell
$env:GENAI_API_KEY = 'your_api_key_here'
```

4. Build the project:

```powershell
go build -v ./...
```

5. Run the server:

```powershell
.\redditor.exe
```

6. Open a browser and visit: http://localhost:8080/

## HTTP endpoints

- `GET /` — Main web UI (`templates/index.html`).
- `GET /analyse?keyword=<kw>&max=<n>` — Start a scrape + AI analysis for `<kw>`. Returns JSON with results. `max` controls max posts (10 default). The frontend calls this.
- `GET /gethistory` — Returns JSON with all historical analysis results stored in MongoDB.
- `GET /history` — Web UI that loads data from `/gethistory` and displays charts.
- Static assets served under `/static/` (mapped to `templates/static`).

Note: `server.go` currently registers handlers using `mux.HandleFunc("GET /static/", ...)` which is non-standard: the standard API is `mux.HandleFunc("/static/", ...)`. The app still works with the registered handlers used here. If you encounter static file 404s, change these lines in `server.go` to use paths without the `GET ` prefix.

## Data flow

1. User provides a keyword on the front page and clicks Search.
2. Frontend hits `/analyse` which calls `scrapers.ScrapeReddit`.
3. Scraper finds subreddits, fetches posts and nested comments, respecting a simple persisted rate limit via `last_request.txt`.
4. Collected posts are marshalled to JSON and passed to `ai.AnalyzePosts`, which calls Google GenAI to extract structured pain points.
5. Results are filtered, saved to `person.json`, and inserted into the `reddit.posts` MongoDB collection.
6. `/gethistory` returns all documents in that collection for dashboards.

## Rate limiting and politeness

The fetcher uses a small persisted rate-limit mechanism (writes last request time to `last_request.txt`) and waits to ensure ~20 requests/sec or slower across runs. Respectful crawling is still your responsibility — consider increasing the interval to be kinder to Reddit.

## Development notes & TODOs

- Move hard-coded API key into environment variables.
- Improve error handling and responses for `/analyse` (currently prints errors server-side and may return empty responses).
- Limit concurrency and add context cancellation propagation to `AnalyseHandler` to support stopping long scrapes.
- Add unit tests for scrapers and AI parsing logic.

## Troubleshooting

- MongoDB connection error: ensure MongoDB is running and reachable at the URI in `server.go`.
- AI errors: ensure a valid GenAI API key and internet connectivity.
- Static files 404: inspect `server.go` handler registration for `/static/` and update to `mux.Handle("/static/", ...)` if necessary.

## License

This repository has no license file. Treat it as private. Add a LICENSE file if you plan to open-source it.

---

If you want, I can:

- Replace the hard-coded API key with reading from an environment variable and update `ai/ai.go` accordingly.
- Run `go build` and fix any build errors.
- Add a small README section showing example output and screenshots.
