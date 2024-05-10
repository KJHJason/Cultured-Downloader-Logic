package pixiv

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api"
)

// PixivDl contains the IDs of the Pixiv artworks and
// illustrators and Tag Names to download.
type PixivDl struct {
	ArtworkIds []string

	ArtistIds      []string
	ArtistPageNums []string

	TagNames         []string
	TagNamesPageNums []string
}

// ValidateArgs validates the IDs of the Pixiv artworks and illustrators to download.
//
// It also validates the page numbers of the tag names to download.
//
// Should be called after initialising the struct.
func (p *PixivDl) ValidateArgs() error {
	err := api.ValidateIds(p.ArtworkIds)
	if err != nil {
		return err
	}

	err = api.ValidateIds(p.ArtistIds)
	if err != nil {
		return err
	}

	p.ArtworkIds = api.RemoveSliceDuplicates(p.ArtworkIds)
	if len(p.ArtistPageNums) > 0 {
		err = api.ValidatePageNumInput(
			len(p.ArtistIds),
			p.ArtistPageNums,
			[]string{
				"Number of illustrators ID(s) and illustrators' page numbers must be equal.",
			},
		)
		if err != nil {
			return err
		}
	} else {
		p.ArtistPageNums = make([]string, len(p.ArtistIds))
	}

	p.ArtistIds, p.ArtistPageNums = api.RemoveDuplicateIdAndPageNum(
		p.ArtistIds,
		p.ArtistPageNums,
	)

	if len(p.TagNamesPageNums) > 0 {
		err = api.ValidatePageNumInput(
			len(p.TagNames),
			p.TagNamesPageNums,
			[]string{
				"Number of tag names and tag names' page numbers must be equal.",
			},
		)
		if err != nil {
			return err
		}
	} else {
		p.TagNamesPageNums = make([]string, len(p.TagNames))
	}

	p.TagNames, p.TagNamesPageNums = api.RemoveDuplicateIdAndPageNum(
		p.TagNames,
		p.TagNamesPageNums,
	)
	return nil
}
