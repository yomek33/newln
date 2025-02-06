package gemini

// func TestRealGeminiClient_GenerateJsonContent(t *testing.T) {
// 	if err := godotenv.Load("./../../../.env"); err != nil {
// 		t.Fatalf("error loading .env file: %v", err)
// 	}

// 	apiKey := os.Getenv("GEMINI_API_KEY")
// 	if apiKey == "" {
// 		t.Fatal("GEMINI_API_KEY is not set")
// 	}

// 	ctx := context.Background()
// 	client, err := NewRealGeminiClient(ctx, apiKey)
// 	if err != nil {
// 		t.Fatalf("failed to create Gemini client: %v", err)
// 	}

// 	prompt := "Generate a JSON array of 3 programming tips"
// 	response, err := client.GenerateJsonContent(ctx, prompt)
// 	if err != nil {
// 		t.Fatalf("API call failed: %v", err)
// 	}

// 	// JSON の構文チェック
// 	var parsedData []map[string]interface{}
// 	if err := json.Unmarshal(response, &parsedData); err != nil {
// 		t.Fatalf("Invalid JSON response: %v", err)
// 	}

// 	// 期待するデータがあるか
// 	if len(parsedData) == 0 {
// 		t.Fatalf("Expected JSON array, but got empty response")
// 	}

// 	t.Logf("Success! Received JSON: %s", string(response))
// }
