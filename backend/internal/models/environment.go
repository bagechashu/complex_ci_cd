package models

import "time"

type Environment struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Rank      int       `json:"rank"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EnvironmentRequest struct {
	Name string `json:"name" binding:"required"`
	Rank int    `json:"rank"`
}
