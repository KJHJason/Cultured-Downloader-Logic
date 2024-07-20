package pixivmobile

import (
	"context"
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// ValidateArgs validates the arguments of the Pixiv download options.
//
// Should be called after initialising the struct.
func (p *PixivMobile) ValidateArgs() error {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	if p.Base.MainProgBar() == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: main progress bar is empty",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.Base.Configs == nil {
		return fmt.Errorf(
			"pixiv mobile error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if err := p.pFilters.ValidateForMobileApi(p.user.IsPremium); err != nil {
		return err
	}

	if p.Base.UseCacheDb && p.Base.Configs.OverwriteFiles {
		p.Base.UseCacheDb = false
	}

	if dlDirPath, err := utils.ValidateDlDirPath(p.Base.DownloadDirPath, constants.PIXIV_MOBILE_TITLE); err != nil {
		return err
	} else {
		p.Base.DownloadDirPath = dlDirPath
	}

	if p.Base.Notifier == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: notifier is nil",
			cdlerrors.DEV_ERROR,
		)
	}
	return nil
}
