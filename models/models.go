package models

import (
	"net/http"

	"github.com/F0RG-2142/capstone-1/internal/database"
)

type apiConfig struct {
	DB       *database.Queries
	Platform string
	Secret   string
}

type Middleware func(http.Handler) http.Handler

var Cfg apiConfig
