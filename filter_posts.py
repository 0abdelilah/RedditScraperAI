import json

# Load JSON from a file
with open("filtered_data.json", "r") as f:
    data = json.load(f)

# Filter posts
for item in data:
    item["posts"] = [
        post for post in item["posts"]
        if post.get("reoccurrence", 0) >= 5 and post.get("classification") in ["tool_limitation", "niche"]
    ]

# Remove keywords with no posts left
data = [item for item in data if item["posts"]]

# Save the filtered JSON
with open("data.json", "w") as f:
    json.dump(data, f, indent=4)
