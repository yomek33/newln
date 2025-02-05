package services

//Geminiからのレスポンス
type PhraseResponse struct {
	ID          int    `json:"id"`
	Collocation string `json:"collocation"`
	FromText    bool   `json:"from_text"`
	Example     string `json:"example"`
	Difficulty  string `json:"difficulty"`
}

