package pixiv

type SearchFilters struct {
	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder  string
	SearchMode string

	// For web api: 1 = filter AI works, 0 = Display AI works
	// For mobile api: 0 = filter AI works, 1 = Display AI works
	SearchAiMode int

	RatingMode  string
	ArtworkType string
}
