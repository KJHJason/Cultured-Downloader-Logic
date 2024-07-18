package metadata

type KemonoPostEmbeddedContent struct {
	Description string `json:"description,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Url         string `json:"url,omitempty"`
}

type KemonoPost struct {
	PostId       string                    `json:"post_id"`
	Url          string                    `json:"url"`
	Title        string                    `json:"title"`
	Service      string                    `json:"service"`
	Content      string                    `json:"content"`
	PublishedUTC string                    `json:"published_utc"`
	EmbedContent KemonoPostEmbeddedContent `json:"embed_content"`
}
