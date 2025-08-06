package models

type UploadResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Files      []FileInfo  `json:"files"`
	Flashcards []Flashcard `json:"flashcards"`
}
