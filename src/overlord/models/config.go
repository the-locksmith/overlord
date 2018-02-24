package models

type Config struct {
	Data    string `json:"data"`
	Debug   bool   `json:"debug"`
	SeedURL string `json:"seed_url"`
}
