package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

type PainPointResponse struct {
	Has_pain_point     bool    `json:"has_pain_point"`
	Summary            string  `json:"summary"`
	Microsaas_solution string  `json:"microsaas_solution"`
	Rating             float64 `json:"rating"`
}

func analyzePostContent(postContent string) PainPointResponse {
	prompt := fmt.Sprintf(`TASK: You are given several Reddit posts (each with title, body, comments, and replies).

For EACH post, extract the following:

1. "pain_point": The main frustration/problem explicitly mentioned (1–2 sentences, quoted or paraphrased directly from the text).
2. "classification": One of:
    "personal" – individual life/behavioral struggles
    "niche" – very specific audience or rare problem
    "tool_limitation" – shortcomings of an existing product/service
    "time_drain" – problem wastes excessive time
3. "problem_type": Short label (e.g., "time_waste", "money_loss", "unclear_process", etc.).
4. "reoccurrence": Integer (count how many times the same/similar pain point appears across posts).

RULES:
- Only use information explicitly present in the posts/comments.
- Do NOT invent, speculate, or generalize beyond what is written.
- Be concise, consistent, and objective.
- Output ONLY valid JSON (no explanations or extra text).
- If no clear pain point exists, return empty strings for fields.
- Only include pain points that people would realistically pay to solve (because it will gain them money or wastes significant time...).

RETURN JSON:
[
  {
    "id": <post_id>,
    "pain_point": "...",
    "classification": "...",
    "problem_type": "...",
    "reoccurrence": <int>
  }
]

Input: %s`, postContent)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: "AIzaSyAvqlxoIQqupWEKO_Hd3OgO-pIdUiljnD0",
	})
	if err != nil {
		log.Fatal(err)
	}

	result, _ := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)

	clean := strings.TrimPrefix(result.Text(), "```json")
	clean = strings.TrimSuffix(clean, "```")

	fmt.Println(clean)
	var painpoint PainPointResponse
	json.Unmarshal([]byte(clean), &painpoint)

	return painpoint
}
