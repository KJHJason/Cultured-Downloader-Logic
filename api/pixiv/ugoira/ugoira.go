package ugoira

import (
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

// UgoiraDlOptions is the struct that contains the
// configs for the processing of the ugoira images after downloading from Pixiv.
type UgoiraOptions struct {
	DeleteZip    bool
	Quality      int
	OutputFormat string
	UseCacheDb   bool
}

var UGOIRA_ACCEPTED_EXT = []string{
	".gif",
	".apng",
	".webp",
	".webm",
	".mp4",
}

// ValidateArgs validates the arguments of the ugoira process options.
//
// Should be called after initialising the struct.
func (u *UgoiraOptions) ValidateArgs() error {
	u.OutputFormat = strings.ToLower(u.OutputFormat)

	// u.Quality is only for .mp4 and .webm
	if u.OutputFormat == ".mp4" && u.Quality < 0 || u.Quality > 51 {
		return fmt.Errorf(
			"pixiv error %d: Ugoira quality of %d is not allowed\nUgoira quality for FFmpeg must be between 0 and 51 for .mp4",
			cdlerrors.INPUT_ERROR,
			u.Quality,
		)
	} else if u.OutputFormat == ".webm" && u.Quality < 0 || u.Quality > 63 {
		return fmt.Errorf(
			"pixiv error %d: Ugoira quality of %d is not allowed\nUgoira quality for FFmpeg must be between 0 and 63 for .webm",
			cdlerrors.INPUT_ERROR,
			u.Quality,
		)
	}

	u.OutputFormat = strings.ToLower(u.OutputFormat)
	_, err := api.ValidateStrArgs(
		u.OutputFormat,
		UGOIRA_ACCEPTED_EXT,
		[]string{
			fmt.Sprintf(
				"pixiv error %d: Output extension %q is not allowed for ugoira conversion",
				cdlerrors.INPUT_ERROR,
				u.OutputFormat,
			),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
