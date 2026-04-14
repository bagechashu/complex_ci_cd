package models

import "time"

type Application struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Repo      string    `json:"repo,omitempty"`
	BuildType string    `json:"build_type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ApplicationRequest struct {
	Name      string `json:"name" binding:"required"`
	Repo      string `json:"repo"`
	BuildType string `json:"build_type"`
}
