package fantia

type FantiaContent struct {
	// Any attachments such as pdfs that are on their dedicated section
	AttachmentURI string `json:"attachment_uri"`

	Category string `json:"category"`

	// For images that are uploaded to their own section
	PostContentPhotos []struct {
		ID  int `json:"id"`
		URL struct {
			Original string `json:"original"`
		} `json:"url"`
	} `json:"post_content_photos,omitempty"`

	// For images that are embedded in the post content blocks.
	// Could also contain links to other external file hosting providers.
	Comment string `json:"comment"`

	// for attachments such as pdfs that are embedded in the post content
	DownloadUri string `json:"download_uri"`
	Filename    string `json:"filename"`
}

type CaptchaResponse struct {
	Redirect string `json:"redirect"` // if get flagged by the system, it will redirect to this recaptcha url
}

type FantiaPost struct {
	Post struct {
		ID       int    `json:"id"`
		Comment  string `json:"comment"` // the main post content
		Title    string `json:"title"`
		PostedAt string `json:"posted_at"` // Wed, 14 Feb 2024 20:00:00 +0900
		Thumb    struct {
			Original string `json:"original"`
		} `json:"thumb"`
		Fanclub struct {
			User struct {
				Name string `json:"name"`
			} `json:"user"`
			FanclubNameWithCreatorName string `json:"fanclub_name_with_creator_name"`
		} `json:"fanclub"`
		Status       string          `json:"status"`
		PostContents []FantiaContent `json:"post_contents"`
	} `json:"post"`
	// Redirect string `json:"redirect"`
}

// found in the head HTML tag.
// Although it's a slice, it should only contains one element.
type ProductInfo []struct {
	Type        string   `json:"@type"`
	Context     string   `json:"@context"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Image       []string `json:"image"`
	Brand       struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"brand"`
	Offers struct {
		Type          string `json:"@type"`
		Price         int    `json:"price"`
		PriceCurrency string `json:"priceCurrency"`
		URL           string `json:"url"`
		Availability  string `json:"availability"`
	} `json:"offers"`
}
