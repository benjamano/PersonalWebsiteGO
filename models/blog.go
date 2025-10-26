package models

import (
	"time"
	"html/template"
)

// Blog represents a blog post
type Blog struct {
	ID        int             `json:"id"`
	Title     string          `json:"title"`
	Content   template.HTML   `json:"content"`
	Author    string          `json:"author"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
