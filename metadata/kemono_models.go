package metadata

type KemonoPostEmbeddedContent struct {
	Description string `json:"description"`
	Subject     string `json:"subject"`
	Url         string `json:"url"`
}

type KemonoPost struct {
	PostId       string                    `json:"post_id"`
	Url          string                    `json:"url"`
	Title        string                    `json:"title"`
	Service      string                    `json:"service"`
	Content      string                    `json:"content"`
	PublishedUTC string                    `json:"published_utc"`
	EmbedContent KemonoPostEmbeddedContent `json:"embed_content";omitempty`
}
