package models

type Flashcard struct {
	ID         string `json:"id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	Difficulty string `json:"difficulty"`
	Source     string `json:"source"`
	CreatedAt  string `json:"created_at"`
}
