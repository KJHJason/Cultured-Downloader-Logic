package kemono

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

type KemonoDl struct {
	CreatorUrls     []string
	CreatorPageNums []string
	CreatorsToDl    []*KemonoCreatorToDl

	PostUrls  []string
	PostsToDl []*KemonoPostToDl

	DlFav bool
}

func ProcessCreatorUrls(creatorUrls []string, pageNums []string) []*KemonoCreatorToDl {
	creatorsToDl := make([]*KemonoCreatorToDl, len(creatorUrls))
	for i, creatorUrl := range creatorUrls {
		matched := constants.KEMONO_CREATOR_URL_REGEX.FindStringSubmatch(creatorUrl)
		creatorsToDl[i] = &KemonoCreatorToDl{
			Service:   matched[constants.KEMONO_CREATOR_URL_REGEX_SERVICE_IDX],
			CreatorId: matched[constants.KEMONO_CREATOR_URL_REGEX_CREATOR_ID_IDX],
			PageNum:   pageNums[i],
		}
	}

	return creatorsToDl
}

func ProcessPostUrls(postUrls []string) []*KemonoPostToDl {
	postsToDl := make([]*KemonoPostToDl, len(postUrls))
	for i, postUrl := range postUrls {
		matched := constants.KEMONO_POST_URL_REGEX.FindStringSubmatch(postUrl)
		postsToDl[i] = &KemonoPostToDl{
			Service:   matched[constants.KEMONO_POST_URL_REGEX_SERVICE_IDX],
			CreatorId: matched[constants.KEMONO_POST_URL_REGEX_CREATOR_ID_IDX],
			PostId:    matched[constants.KEMONO_POST_URL_REGEX_POST_ID_IDX],
		}
	}

	return postsToDl
}

// RemoveDuplicates removes duplicate creators and posts from the slice
func (k *KemonoDl) RemoveDuplicates() {
	if len(k.CreatorsToDl) > 0 {
		newCreatorSlice := make([]*KemonoCreatorToDl, 0, len(k.CreatorsToDl))
		seen := make(map[string]struct{})
		for _, creator := range k.CreatorsToDl {
			key := fmt.Sprintf("%s/%s", creator.Service, creator.CreatorId)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			newCreatorSlice = append(newCreatorSlice, creator)
		}
		k.CreatorsToDl = newCreatorSlice
	}

	if len(k.PostsToDl) == 0 {
		return
	}
	newPostSlice := make([]*KemonoPostToDl, 0, len(k.PostsToDl))
	seen := make(map[string]struct{})
	for _, post := range k.PostsToDl {
		key := fmt.Sprintf("%s/%s/%s", post.Service, post.CreatorId, post.PostId)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		newPostSlice = append(newPostSlice, post)
	}
	k.PostsToDl = newPostSlice
}

func (k *KemonoDl) ValidateArgs() error {
	valid, outlier := utils.SliceMatchesRegex(constants.KEMONO_CREATOR_URL_REGEX, k.CreatorUrls, true)
	if !valid {
		return fmt.Errorf(
			"kemono error %d: invalid creator URL found for kemono: %s",
			cdlerrors.INPUT_ERROR,
			outlier,
		)
	}

	valid, outlier = utils.SliceMatchesRegex(constants.KEMONO_POST_URL_REGEX, k.PostUrls, true)
	if !valid {
		return fmt.Errorf(
			fmt.Sprintf(
				"kemono error %d: invalid post URL found for kemono: %s",
				cdlerrors.INPUT_ERROR,
				outlier,
			),
		)
	}

	if len(k.CreatorUrls) > 0 {
		if len(k.CreatorPageNums) == 0 {
			k.CreatorPageNums = make([]string, len(k.CreatorUrls))
		} else {
			err := utils.ValidatePageNumInput(
				len(k.CreatorUrls),
				k.CreatorPageNums,
				[]string{
					"Number of creator URL(s) and page numbers must be equal.",
				},
			)
			if err != nil {
				return err
			}
		}
		creatorsToDl := ProcessCreatorUrls(k.CreatorUrls, k.CreatorPageNums)
		k.CreatorsToDl = append(k.CreatorsToDl, creatorsToDl...)
		k.CreatorUrls = nil
		k.CreatorPageNums = nil
	}
	if len(k.PostUrls) > 0 {
		postsToDl := ProcessPostUrls(k.PostUrls)
		k.PostsToDl = append(k.PostsToDl, postsToDl...)
		k.PostUrls = nil
	}
	k.RemoveDuplicates()
	return nil
}

// KemonoDlOptions is the struct that contains the arguments for Kemono download options.
type KemonoDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc
	Base   *api.BaseDl // Note: Base.DlAttachments is not just .psd files but also the post content images.
}

func (k *KemonoDlOptions) GetContext() context.Context {
	return k.ctx
}

func (k *KemonoDlOptions) SetContext(ctx context.Context) {
	k.ctx, k.cancel = context.WithCancel(ctx)
}

// CancelCtx releases the resources used and cancels the context of the KemonoDlOptions struct.
func (k *KemonoDlOptions) CancelCtx() {
	k.cancel()
}

func (k *KemonoDlOptions) CtxIsActive() bool {
	return k.ctx.Err() == nil
}

// ValidateArgs validates the session cookie ID of the Kemono account to download from.
// It also validates the Google Drive client if the user wants to download to Google Drive.
//
// Should be called after initialising the struct.
func (k *KemonoDlOptions) ValidateArgs(userAgent string) error {
	if k.GetContext() == nil {
		k.SetContext(context.Background())
	}

	if k.Base.Notifier == nil {
		return fmt.Errorf(
			"kemono error %d, notifier is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if k.Base.MainProgBar() == nil {
		return fmt.Errorf(
			"kemono error %d, main progress bar is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if k.Base.Configs == nil {
		return fmt.Errorf(
			"kemono error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if k.Base.UseCacheDb && k.Base.Configs.OverwriteFiles {
		k.Base.UseCacheDb = false
	}

	if dlDirPath, err := utils.ValidateDlDirPath(k.Base.DownloadDirPath, constants.KEMONO_TITLE); err != nil {
		return err
	} else {
		k.Base.DownloadDirPath = dlDirPath
	}

	if k.Base.SessionCookieId != "" {
		k.Base.SessionCookies = []*http.Cookie{
			api.GetCookie(k.Base.SessionCookieId, constants.KEMONO),
		}
		k.Base.SessionCookieId = ""
	}

	if len(k.Base.SessionCookies) > 0 {
		ch := httpfuncs.CaptchaHandler{}
		if err := api.VerifyCookies(constants.KEMONO, userAgent, k.Base.SessionCookies, ch); err != nil {
			return err
		}
	} else {
		return fmt.Errorf(
			"kemono error %d: session cookie is required",
			cdlerrors.INPUT_ERROR,
		)
	}

	if k.Base.DlGdrive && k.Base.GdriveClient == nil {
		k.Base.DlGdrive = false
	} else if !k.Base.DlGdrive && k.Base.GdriveClient != nil {
		k.Base.GdriveClient = nil
	}
	return nil
}
