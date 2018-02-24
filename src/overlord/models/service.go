package models

type Service struct {
	Description string `json:"description"`
	Port        int    `json:"port"`
}
