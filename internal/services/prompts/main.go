package main

import (
	"context"
	"strings"

	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

const model = "gemini-2.0-flash-exp"

// プロンプトを試す
func main() {
	fileFlag := flag.String("p", "Your prompt text here", "The prompt text to generate content")
	flag.Parse()
	if *fileFlag == "" {
		fmt.Println("⚠️  使用するプロンプトファイルを -p で指定")
		os.Exit(1)
	}

	// ファイルを読み込む
	promptFile, err := os.ReadFile(*fileFlag)
	if err != nil {
		fmt.Printf("❌ ファイル %s の読み込みエラー: %v\n", *fileFlag, err)
		os.Exit(1)
	}
	prompt := string(promptFile)

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("failed to create genai client: %v", err)
	}

	geminiConfig := genai.GenerateContentConfig{
		MaxOutputTokens:  genai.Ptr(int64(8192)),
		TopK:             genai.Ptr(float64(40)),
		TopP:             genai.Ptr(0.95),
		Temperature:      genai.Ptr(float64(1)),
		ResponseMIMEType: "application/json",
	}

	switch *fileFlag {
	case "generate_phrase.txt":
		respSchema := &genai.Schema{
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"id": {
						Type: genai.TypeInteger,
					},
					"collocation": {
						Type: genai.TypeString,
					},
					"from_text": {
						Type: genai.TypeBoolean,
					},
					"example": {
						Type: genai.TypeString,
					},
					"difficulty": {
						Type: genai.TypeString,
					},
				},
				Required: []string{"id", "collocation", "from_text", "example", "difficulty"},
			},
		}
		prompt = strings.ReplaceAll(prompt, "{{TEXT}}", sampleText)
		geminiConfig.ResponseMIMEType = "application/json"
		geminiConfig.ResponseSchema = respSchema

	}
	res, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", genai.Text(prompt), &geminiConfig)
	if err != nil {
		log.Fatalf("failed to generate content: %v", err)
	}

	fmt.Printf("Response: %+v\n", res)

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		log.Fatalf("no content generated")
	}
	printResponse(res)
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}

const sampleText = "[WASHINGTON — Booz Allen Hamilton deployed a generative AI large language model on the International Space Station using a Hewlett Packard Enterprise advanced edge computer designed for in-orbit experiments.\n" +
	"The generative AI large language model (LLM) has been in operation since mid-July as part of an experiment, Booz Allen announced Aug. 1.\n" +
	"A generative AI large language model is a sophisticated type of artificial intelligence designed to understand and generate human language. The LLM at the space station is intended to help astronauts address queries and resolve issues.\n" +
	"“Right now, astronauts train for many hours to be able to conduct repairs of machinery and onboard systems. However, having the ability to ask the instruction manuals questions and receive relevant and rapid responses could augment their efforts so they can fix problems at an accelerated pace,” said Dan Wald, principal AI solutions architect for space applications at Booz Allen.\n" +
	"The Hewlett Packard Enterprise (HPE) Spaceborne Computer-2, launched in February 2021, provides the infrastructure for advanced experiments, including AI and machine learning in space. By processing data in orbit and sending only the insights back to Earth, it reduces data transmission times.\n" +
	"Spaceborne Computer-2 has completed multiple research experiments in fields such as DNA sequencing, image processing, natural disaster recovery, 3D printing and 5G technology.\n" +
	"“Generative AI in space is truly the new frontier,” said Chris Bogdan, executive vice president at Booz Allen and leader of the firm’s space business. “Booz Allen is committed to pushing the boundaries of what is possible with AI and other mission-critical technologies in space.”]"
