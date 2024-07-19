package metadata

import (
	"time"
)

type FantiaPost struct {
	Url                  string    `json:"url"`
	PostedAt             time.Time `json:"datetime"`
	Title                string    `json:"title"`
	PostComment          string    `json:"post_comment"`
	EmbeddedPostComments []string  `json:"embedded_post_comments"`
}

type FantiaProductPricing struct {
	Price    int    `json:"price"`
	Currency string `json:"price_currency"`
}

type FantiaProduct struct {
	Url         string               `json:"url"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Images      []string             `json:"images"`
	Pricing     FantiaProductPricing `json:"Pricing"`
}
