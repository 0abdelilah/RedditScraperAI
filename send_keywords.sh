#!/bin/bash

# List of keywords (subreddit names or phrases, spaces allowed)
KEYWORDS=(
    "investing"
    "personal finance"
    "stocks"
    "cryptocurrency"
    "financialindependence"
    "economics"
    "daytrading"
    "options"
    "forex"
    "realestate"
    "budgeting"
    "passiveincome"
    "retirement"
    "wealthmanagement"
    "stockmarket"
    "valueinvesting"
    "cryptomarkets"
    "startups"
    "business"
    "sidehustle"
)

# Max posts per request
MAX_POSTS=10

# Infinite loop
while true; do
    for keyword in "${KEYWORDS[@]}"; do
        echo "Requesting keyword: $keyword"

        # URL encode spaces as %20
        encoded_keyword=$(echo "$keyword" | jq -s -R -r @uri)

        # Send request
        curl -s "http://localhost:8080/analyse?keyword=$encoded_keyword&maxPosts=$MAX_POSTS"

        echo -e "\nFinished request for '$keyword', waiting 1s before next..."
        sleep 1
    done
done
