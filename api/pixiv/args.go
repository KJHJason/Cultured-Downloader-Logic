package pixiv

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api"
)

// PixivDl contains the IDs of the Pixiv artworks and
// illustrators and Tag Names to download.
type PixivDl struct {
	ArtworkIds []string

	IllustratorIds      []string
	IllustratorPageNums []string

	TagNames         []string
	TagNamesPageNums []string
}

// ValidateArgs validates the IDs of the Pixiv artworks and illustrators to download.
//
// It also validates the page numbers of the tag names to download.
//
// Should be called after initialising the struct.
func (p *PixivDl) ValidateArgs() {
	api.ValidateIds(p.ArtworkIds)
	api.ValidateIds(p.IllustratorIds)
	p.ArtworkIds = api.RemoveSliceDuplicates(p.ArtworkIds)

	if len(p.IllustratorPageNums) > 0 {
		api.ValidatePageNumInput(
			len(p.IllustratorIds),
			p.IllustratorPageNums,
			[]string{
				"Number of illustrators ID(s) and illustrators' page numbers must be equal.",
			},
		)
	} else {
		p.IllustratorPageNums = make([]string, len(p.IllustratorIds))
	}
	p.IllustratorIds, p.IllustratorPageNums = api.RemoveDuplicateIdAndPageNum(
		p.IllustratorIds,
		p.IllustratorPageNums,
	)

	if len(p.TagNamesPageNums) > 0 {
		api.ValidatePageNumInput(
			len(p.TagNames),
			p.TagNamesPageNums,
			[]string{
				"Number of tag names and tag names' page numbers must be equal.",
			},
		)
	} else {
		p.TagNamesPageNums = make([]string, len(p.TagNames))
	}
	p.TagNames, p.TagNamesPageNums = api.RemoveDuplicateIdAndPageNum(
		p.TagNames,
		p.TagNamesPageNums,
	)
}
