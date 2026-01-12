package model

type Script struct {
	Name        string `json:"name"`
	Command     string `json:"command"`
	Description string `json:"description"`
}
