package metadata

import (
	"time"
)

type PixivFanboxPost struct {
	PostUrl            string    `json:"post_url"`
	Title              string    `json:"title"`
	PublishedAt        time.Time `json:"published_at"`
	HasAdultContent    bool      `json:"has_adult_content"`
	RestrictedFromUser bool      `json:"restricted_from_user"`
	PostType           string    `json:"post_type"`
	PlanFee            int       `json:"plan_fee"`
}
