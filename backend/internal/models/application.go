package models

import (
	"fmt"
	"time"
)

type Application struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	ImageName   string     `db:"image_name" json:"image_name"`
	GitRepo     *string    `db:"git_repo" json:"git_repo"`
	BuildType   *string    `db:"build_type" json:"build_type"`
	Description *string    `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

func (a *Application) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("application name is required")
	}
	if a.ImageName == "" {
		return fmt.Errorf("image_name is required")
	}
	return nil
}

func (a *Application) GetID() interface{} {
	return a.ID
}
