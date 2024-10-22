package fantia

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/metadata"
	"github.com/PuerkitoBio/goquery"
)

type postContentId struct {
	postId    int
	commentId int
}

func dlImagesFromPost(content *FantiaContent, postFolderPath string, organise bool, id *postContentId) []*httpfuncs.ToDownload {
	postContentPhotos := content.PostContentPhotos
	matchedUrlInComments := constants.FANTIA_COMMENT_IMAGE_URL_REGEX.FindAllStringSubmatch(content.Comment, -1)
	commentsLen := len(matchedUrlInComments)
	postContentLen := len(postContentPhotos)
	if commentsLen == 0 && postContentLen == 0 {
		return make([]*httpfuncs.ToDownload, 0)
	}
	urlsSlice := make([]*httpfuncs.ToDownload, 0, commentsLen+postContentLen)

	// for images that are embedded in the post content
	commentCount := 0
	commentFolderId := strconv.Itoa(id.commentId)
	for _, matched := range matchedUrlInComments {
		// Note: urls in the comments are now absolute paths, no need to add constants.FANTIA_URL prefix.
		imageUrl := matched[constants.FANTIA_COMMENT_REGEX_URL_IDX]

		// Since the API returns the URL with escaped & characters, we need to replace them so that the URL is valid.
		imageUrl = strings.Replace(imageUrl, "\\u0026", "&", 2)

		filePath := filepath.Join(postFolderPath, constants.FANTIA_POST_BLOG_DIR_NAME)
		if organise {
			commentCount++
			fileExt := matched[constants.FANTIA_COMMENT_REGEX_EXT_IDX]
			filePath = filepath.Join(filePath, commentFolderId, fmt.Sprintf("%d.%s", commentCount, fileExt))
		}

		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      imageUrl,
			FilePath: filePath,
		})
	}
	if organise && commentsLen > 0 {
		id.commentId++
	}

	// download images that are uploaded to their own section
	postCount := 0
	postFolderId := strconv.Itoa(id.postId)
	for _, image := range postContentPhotos {
		imageUrl := image.URL.Original
		filePath := filepath.Join(postFolderPath, constants.IMAGES_FOLDER)

		if organise {
			postCount++
			matched := constants.FANTIA_IMAGE_URL_REGEX.FindStringSubmatch(imageUrl)
			if len(matched) > 0 {
				fileExt := matched[constants.FANTIA_IMAGE_URL_REGEX_EXT_IDX]
				filePath = filepath.Join(filePath, postFolderId, fmt.Sprintf("%d.%s", postCount, fileExt))
			} else {
				err := fmt.Errorf(
					"fantia error %d: failed to match image url %q when trying to organise images",
					cdlerrors.UNEXPECTED_ERROR,
					imageUrl,
				)
				logger.LogError(err, logger.ERROR)
			}
		}

		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      imageUrl,
			FilePath: filePath,
		})
	}
	if organise && postContentLen > 0 {
		id.postId++
	}
	return urlsSlice
}

func dlAttachmentsFromPost(content *FantiaContent, postFolderPath string) []*httpfuncs.ToDownload {
	var urlsSlice []*httpfuncs.ToDownload

	// get the attachment url string if it exists
	attachmentUrl := content.AttachmentURI
	if attachmentUrl != "" {
		attachmentUrlStr := constants.FANTIA_URL + attachmentUrl
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      attachmentUrlStr,
			FilePath: filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER),
		})
	} else if content.DownloadUri != "" {
		// if the attachment url string does not exist,
		// then get the download url for the file
		downloadUrl := constants.FANTIA_URL + content.DownloadUri
		filename := content.Filename
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      downloadUrl,
			FilePath: filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER, filename),
		})
	}
	return urlsSlice
}

// Convert string value like "Wed, 14 Feb 2024 20:00:00 +0900" to time.Time
func parseDateStrToDateTime(dateStr string) time.Time {
	dateTime, err := time.Parse(time.RFC1123Z, dateStr)
	if err != nil {
		errMsg := fmt.Errorf(
			"fantia error %d: failed to parse date string %q to datetime: %w",
			cdlerrors.UNEXPECTED_ERROR,
			dateStr,
			err,
		)
		logger.LogError(errMsg, logger.ERROR)
		return time.Time{}
	}
	return dateTime
}

// Process the JSON response from Fantia's API and
// returns a slice of urls and a slice of gdrive urls to download from
func processFantiaPost(res *httpfuncs.ResponseWrapper, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	respBody, err := res.GetBody()
	if err != nil {
		return nil, nil, err
	}
	resUrl := res.Url()

	// processes a fantia post
	// returns a map containing the post id and the url to download the file from
	var postJson FantiaPost
	if err := httpfuncs.LoadJsonFromBytes(resUrl, respBody, &postJson); err != nil {
		return nil, nil, err
	}

	post := postJson.Post
	postDate := parseDateStrToDateTime(post.PostedAt)
	if !dlOptions.Base.Filters.IsPostDateValid(postDate) {
		return nil, nil, nil
	}
	postId := strconv.Itoa(post.ID)
	postTitle := post.Title
	fanclubName := post.Fanclub.FanclubNameWithCreatorName
	if fanclubName == "" { // just in case but shouldn't happen
		fanclubName = post.Fanclub.User.Name
	}
	postFolderPath := iofuncs.GetPostFolder(dlOptions.Base.DownloadDirPath, fanclubName, postId, postTitle)

	var urlsSlice []*httpfuncs.ToDownload
	thumbnail := post.Thumb.Original
	if dlOptions.Base.DlThumbnails && thumbnail != "" {
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      thumbnail,
			FilePath: postFolderPath,
		})
	}

	gdriveLinks := gdrive.ProcessPostText(
		post.Comment,
		postFolderPath,
		dlOptions.Base.DlGdrive,
		dlOptions.Base.Configs.LogUrls,
	)

	postContent := post.PostContents
	if postContent == nil {
		return urlsSlice, gdriveLinks, nil
	}

	if dlOptions.Base.SetMetadata {
		comments := make([]string, 0, len(postContent))
		for _, content := range postContent {
			comments = append(comments, content.Comment)
		}
		postMetadata := metadata.FantiaPost{
			Url:                  resUrl,
			PostedAt:             postDate,
			Title:                postTitle,
			PostComment:          post.Comment,
			EmbeddedPostComments: comments,
		}
		if err := metadata.WriteMetadata(postMetadata, postFolderPath); err != nil {
			return nil, nil, err
		}
	}

	contentIds := &postContentId{
		commentId: 1,
		postId:    1,
	}
	for _, content := range postContent {
		commentGdriveLinks := gdrive.ProcessPostText(
			content.Comment,
			postFolderPath,
			dlOptions.Base.DlGdrive,
			dlOptions.Base.Configs.LogUrls,
		)
		if len(commentGdriveLinks) > 0 {
			gdriveLinks = append(gdriveLinks, commentGdriveLinks...)
		}
		if dlOptions.Base.DlImages {
			urlsSlice = append(urlsSlice, dlImagesFromPost(&content, postFolderPath, dlOptions.Base.OrganiseImages, contentIds)...)
		}
		if dlOptions.Base.DlAttachments {
			urlsSlice = append(urlsSlice, dlAttachmentsFromPost(&content, postFolderPath)...)
		}
	}
	return urlsSlice, gdriveLinks, nil
}

type processIllustArgs struct {
	respWrapper *httpfuncs.ResponseWrapper
	postId      string
	postIdsLen  int
	msgSuffix   string
}

// Process the JSON response to get the urls to download
func processIllustDetailApiRes(illustArgs *processIllustArgs, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	progress := dlOptions.Base.MainProgBar()
	progress.SetToSpinner()
	progress.UpdateBaseMsg(
		fmt.Sprintf(
			"Processing retrieved JSON for post %s from Fantia %s...",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished processing retrieved JSON for post %s from Fantia %s!",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while processing retrieved JSON for post %s from Fantia %s.\nPlease refer to the logs for more details.",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.Start()
	defer progress.SnapshotTask()

	urlsToDownload, gdriveLinks, err := processFantiaPost(
		illustArgs.respWrapper,
		dlOptions,
	)
	if err != nil {
		progress.Stop(true)
		return nil, nil, err
	}
	progress.Stop(false)
	return urlsToDownload, gdriveLinks, nil
}

func getAndProcessProductPaidContent(purchaseRelativeUrl, productId string, dlOptions *FantiaDlOptions) ([]string, error) {
	respWrapper, err := getFantiaProductPaidContent(purchaseRelativeUrl, productId, dlOptions)
	//lint:ignore SA5001 Ignore the error check from closing the response body
	defer respWrapper.Close()
	if err != nil {
		return nil, err
	}

	respBody, err := respWrapper.GetBodyReader()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		err = fmt.Errorf(
			"fantia error %d: failed to parse response body when getting paid content from Fantia: %w",
			cdlerrors.HTML_ERROR,
			err,
		)
		logger.LogError(err, logger.ERROR)
		return nil, nil
	}

	paidContentUrls := make([]string, 0)
	// get all divs with the class "row row-packed"
	doc.Find("div.row.row-packed").Each(func(i int, s *goquery.Selection) {
		// find the anchor tag with the class "module-thumbnail"
		productUrl, exists := s.Find("a.module-thumbnail").Attr("href")
		if !exists {
			return
		}

		// since an order can have multiple products, check if the product
		// id in the url matches the product id we are looking for to download
		docProductId := httpfuncs.GetLastPartOfUrl(productUrl)
		if docProductId != productId {
			return
		}

		s.Find("a.btn.btn-primary").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.HasPrefix(href, "/products") && strings.HasSuffix(href, "/content_download") {
				paidContentUrls = append(paidContentUrls, constants.FANTIA_URL+href)
			}
		})
	})
	return paidContentUrls, nil
}

func getFanclubNameFromProductPage(productId string, doc *goquery.Document) string {
	fanclubName := doc.Find(".fanclub-show-header h1.fanclub-name a").Text()
	if fanclubName == "" {
		htmlContent, err := doc.Html()
		if err != nil {
			htmlContent = "failed to get HTML"
		}
		//lint:ignore ST1005 Since the html content is long, it's better to have it on a new line for readability
		errMsg := fmt.Errorf(
			"fantia error %d: failed to get fanclub name from product id %q, please report this issue with the html content below;\n%s\n",
			cdlerrors.HTML_ERROR,
			productId,
			htmlContent,
		)
		logger.LogError(errMsg, logger.ERROR)
		fanclubName = constants.FANTIA_UNKNOWN_CREATOR
	}
	return fanclubName
}

func getProductPaidContent(productId string, doc *goquery.Document, dlOptions *FantiaDlOptions) ([]string, error) {
	var purchaseRelativeUrl string

	// Could have just used .First() but for future-proofing, we'll use .Each() and check the Text() content.
	doc.Find("a.alert-link").Each(func(i int, s *goquery.Selection) {
		if purchaseRelativeUrl != "" {
			return
		}
		if s.Text() != "注文詳細・商品ダウンロード" {
			return
		}
		if href, exists := s.Attr("href"); exists && purchaseRelativeUrl == "" {
			purchaseRelativeUrl = href
		}
	})

	if purchaseRelativeUrl == "" { // not purchased
		return make([]string, 0), nil
	}
	return getAndProcessProductPaidContent(purchaseRelativeUrl, productId, dlOptions)
}

type productDetails struct {
	productInfo        ProductInfo
	productName        string
	thumbnailUrl       string
	previewContentUrls []string
}

func getProductDetails(productId string, doc *goquery.Document) productDetails {
	var pd productDetails
	jsonContent := doc.Find("head script[type='application/ld+json']").Text()
	if jsonContent == "" {
		logger.LogError(
			fmt.Errorf(
				"fantia error %d: failed to get product details from product id %q",
				cdlerrors.HTML_ERROR,
				productId,
			),
			logger.ERROR,
		)
		return pd
	}

	// get the product details from the JSON content
	var productInfoSlice []ProductInfo
	if err := json.Unmarshal([]byte(jsonContent), &productInfoSlice); err != nil {
		logger.LogError(
			//lint:ignore ST1005 Since the json content is long, it's better to have it on a new line for readability
			fmt.Errorf(
				"fantia error %d: failed to unmarshal product details from product id %q. More info => %w\nJSON content: %s\n",
				cdlerrors.JSON_ERROR,
				productId,
				err,
				jsonContent,
			),
			logger.ERROR,
		)
		return pd
	}

	if len(productInfoSlice) == 0 {
		logger.LogError(
			fmt.Errorf(
				"fantia error %d: although unmarshalled successfully, there is no element in the product info slice from product id %q",
				cdlerrors.HTML_ERROR,
				productId,
			),
			logger.ERROR,
		)
		return productDetails{}
	}

	product := productInfoSlice[0]
	pd.productInfo = product
	pd.productName = product.Name

	// alternatively, we can use
	// doc.Find(".product-gallery img").Each(func(i int, s *goquery.Selection) { ... }) if this no longer works
	images := product.Image
	if len(images) == 0 {
		return pd
	}

	pd.thumbnailUrl = images[0] // thumbnail doesn't have the prefix and defaults to the main image
	if len(images) > 1 {
		// images here are micro images, hence we need to replace the
		// micro_ filename prefix with the main_ prefix to get the full image
		previewContentUrls := images[1:]
		for i, img := range previewContentUrls {
			previewContentUrls[i] = strings.Replace(img, "micro_", "main_", 1)
		}
		pd.previewContentUrls = previewContentUrls
	}
	return pd
}

// Note: response body is closed in this function
// errors returned are usually due to parsing error or context cancellation
func processProductPage(cacheKey, productId string, dlOptions *FantiaDlOptions, respWrapper *httpfuncs.ResponseWrapper) ([]*httpfuncs.ToDownload, error) {
	defer respWrapper.Close()
	respBody, err := respWrapper.GetBodyReader()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return nil, fmt.Errorf(
			"fantia error %d: failed to parse response body when getting product page from Fantia: %w",
			cdlerrors.HTML_ERROR,
			err,
		)
	}

	pd := getProductDetails(productId, doc)
	fanclubName := getFanclubNameFromProductPage(productId, doc)

	// Check if the user has purchased the product so that we can get and download the paid content as well.
	paidContent, paidContentErr := getProductPaidContent(productId, doc, dlOptions)

	if paidContentErr != nil {
		if errors.Is(paidContentErr, context.Canceled) {
			return nil, paidContentErr
		}
		logger.LogError(paidContentErr, logger.ERROR)
	}

	numOfEl := len(pd.previewContentUrls) + len(paidContent)
	if pd.thumbnailUrl != "" {
		numOfEl++
	}

	toDownload := make([]*httpfuncs.ToDownload, 0, numOfEl)
	dirPath := iofuncs.GetPostFolder(dlOptions.Base.DownloadDirPath, fanclubName, productId, pd.productName)
	dirPath = filepath.Join(
		filepath.Dir(dirPath), // go up one directory
		constants.FANTIA_PRODUCT_DIR_NAME,
		filepath.Base(dirPath), // go back to the original directory
	)

	if dlOptions.Base.SetMetadata {
		productMetadata := metadata.FantiaProduct{
			Url:         constants.FANTIA_PRODUCT_URL + productId,
			Name:        pd.productName,
			Description: pd.productInfo.Description,
			Images:      pd.productInfo.Image,
			Pricing: metadata.FantiaProductPricing{
				Price:    pd.productInfo.Offers.Price,
				Currency: pd.productInfo.Offers.PriceCurrency,
			},
		}
		if err := metadata.WriteMetadata(productMetadata, dirPath); err != nil {
			return nil, err
		}
	}

	if pd.thumbnailUrl != "" {
		toDownload = append(toDownload, &httpfuncs.ToDownload{
			Url:      pd.thumbnailUrl,
			FilePath: dirPath,
			CacheKey: cacheKey,
			CacheFn:  database.CachePost,
		})
	}
	for i, url := range pd.previewContentUrls {
		dlFilePath := filepath.Join(dirPath, constants.FANTIA_PRODUCT_PREVIEW_DIR_NAME)
		if dlOptions.Base.OrganiseImages {
			fileExt := filepath.Ext(url)
			dlFilePath = filepath.Join(dlFilePath, fmt.Sprintf("%d%s", i+1, fileExt))
		}
		toDownload = append(toDownload, &httpfuncs.ToDownload{
			Url:      url,
			FilePath: dlFilePath,
			CacheKey: cacheKey,
			CacheFn:  database.CachePost,
		})
	}
	for _, url := range paidContent {
		toDownload = append(toDownload, &httpfuncs.ToDownload{
			Url:      url,
			FilePath: filepath.Join(dirPath, constants.FANTIA_PRODUCT_PAID_DIR_NAME),
			CacheKey: cacheKey,
			CacheFn:  database.CachePost,
		})
	}
	return toDownload, nil
}
