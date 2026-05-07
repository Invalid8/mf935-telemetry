package main

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Password string `json:"password"`
}

func loadConfig() (Config, error) {
	f, err := os.Open("config.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
