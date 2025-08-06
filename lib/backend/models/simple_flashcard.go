package models

type SimpleFlashcard struct {
	Term		string	`json:term`
	Definition	string	`json:definition`
}