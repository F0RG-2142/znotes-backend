package models

import "github.com/F0RG-2142/capstone-1/internal/database"

type apiConfig struct {
	DB       *database.Queries
	Platform string
	Secret   string
}

var Cfg apiConfig
