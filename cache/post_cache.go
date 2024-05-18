package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

// example of the Key-Value pairs in the database
// |------------|-----------|
// | key_suffix | unix_time |
// |------------|-----------|
// Where key_suffix can be <url>_<suffix> or <post_id>_<suffix>

const (
	SUFFIX          = "|"
	PLATFORM_SUFFIX = "`"
	POST            = SUFFIX + "post"
	UGOIRA          = SUFFIX + "ugoira"
	GDRIVE          = SUFFIX + "gdrive"
	KEMONO_CREATOR  = SUFFIX + "kemonocreator"
)

func ParsePostKey(url, platform string) string {
	return url + PLATFORM_SUFFIX + platform + POST
}

func parseKey(key, category string) string {
	switch category {
	case UGOIRA:
		return key + UGOIRA
	case GDRIVE:
		return key + GDRIVE
	case KEMONO_CREATOR:
		return key + KEMONO_CREATOR
	default:
		return key
	}
}

func SeparatePostKey(keyBytes []byte) (url string, platform string) {
	key := string(keyBytes)

	// key is in the format <url>`<platform>|post
	if !strings.HasSuffix(key, POST) {
		return "", ""
	}

	// Remove the |post suffix
	key = key[:len(key)-len(POST)]

	// split the key into <url> and <platform>
	splitKey := strings.Split(key, PLATFORM_SUFFIX)
	splitLen := len(splitKey)
	if splitLen < 2 {
		// shouldn't happen but just in case
		return "", ""
	}

	platform = splitKey[splitLen-1]
	if splitLen == 2 {
		return splitKey[0], platform
	}

	// shouldn't have an edge case where the url
	// contains the platform suffix but just in case
	url = strings.Join(splitKey[:splitLen-1], PLATFORM_SUFFIX)
	return url, platform
}

func PostCacheExists(key, platform string) bool {
	return len(Get(ParsePostKey(key, platform))) > 0
}

func GDriveCacheExists(key string) bool {
	return len(Get(parseKey(key, GDRIVE))) > 0
}

func UgoiraCacheExists(key string) bool {
	return len(Get(parseKey(key, UGOIRA))) > 0
}

func GetKemonoCreatorCache(key string) string {
	return GetString(parseKey(key, KEMONO_CREATOR))
}

// Note: the setter functions below do not handle errors and will continue as usual

func CachePost(parsedKey string) {
	SetTime(parsedKey, time.Now())
}

func CacheGDrive(key string) {
	SetTime(parseKey(key, GDRIVE), time.Now())
}

func CacheUgoira(key string) {
	SetTime(parseKey(key, UGOIRA), time.Now())
}

func CacheKemonoCreatorName(key string, creatorName string) {
	SetString(parseKey(key, KEMONO_CREATOR), creatorName)
}

type PostCache struct {
	Url      string
	Platform string
	Datetime time.Time
	CacheKey string
}

// Returns a readable format of the website name for the user
//
// Unlike GetReadableSiteStr, this function will return an empty string if the site string doesn't match one of its cases.
func GetReadableSiteStrSafely(site string) string {
	switch site {
	case constants.FANTIA:
		return constants.FANTIA_TITLE
	case constants.PIXIV_FANBOX:
		return constants.PIXIV_FANBOX_TITLE
	case constants.PIXIV:
		return constants.PIXIV_TITLE
	case constants.KEMONO:
		return constants.KEMONO_TITLE
	default:
		return ""
	}
}

// Returns a readable format of the website name for the user
//
// Will panic if the site string doesn't match one of its cases.
func GetReadableSiteStr(site string) string {
	if readableSite := GetReadableSiteStrSafely(site); readableSite != "" {
		return readableSite
	} else {
		// panic since this is a dev error
		panic(
			fmt.Errorf(
				"error %d: invalid website, %q, in GetReadableSiteStr",
				cdlerrors.DEV_ERROR,
				site,
			),
		)
	}
}

func GetAllCacheForPlatform(ctx context.Context, platforms ...string) []*PostCache {
	keys := CacheDb.GetCacheKeyValue(ctx, func(key, _ []byte) bool {
		url, keyPlatform := SeparatePostKey(key)
		if url == "" || keyPlatform == "" {
			return false
		}
		for _, p := range platforms {
			if keyPlatform == p {
				return true
			}
		}
		return false
	})
	if len(keys) == 0 {
		return make([]*PostCache, 0)
	}

	caches := make([]*PostCache, len(keys))
	for i, key := range keys {
		url, platform := SeparatePostKey(key.Key)
		caches[i] = &PostCache{
			Url:      url,
			Platform: GetReadableSiteStrSafely(platform),
			Datetime: ParseBytesToDateTime(key.Val),
			CacheKey: key.GetKey(),
		}
	}
	return caches
}

var platforms = [...]string{
	constants.FANTIA,
	constants.PIXIV,
	constants.PIXIV_FANBOX,
	constants.KEMONO,
}

func GetAllCacheForAllPlatforms(ctx context.Context) []*PostCache {
	return GetAllCacheForPlatform(ctx, platforms[:]...)
}

func DeletePostCacheForAllPlatforms(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		_, keyPlatform := SeparatePostKey(key)
		if keyPlatform == "" {
			return true
		}
		for _, platform := range platforms {
			if keyPlatform == platform {
				return false
			}
		}
		return true
	})
}

func GetAllGdriveCache(ctx context.Context) []*CacheKeyValue {
	return CacheDb.GetCacheKeyValue(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), GDRIVE)
	})
}

func DeleteAllGdriveCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), GDRIVE)
	})
}

func GetAllUgoiraCache(ctx context.Context) []*CacheKeyValue {
	return CacheDb.GetCacheKeyValue(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), UGOIRA)
	})
}

func DeleteAllUgoiraCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), UGOIRA)
	})
}

func GetAllKemonoCreatorCache(ctx context.Context) []*CacheKeyValue {
	return CacheDb.GetCacheKeyValue(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), KEMONO_CREATOR)
	})
}

func DeleteAllKemonoCreatorCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), KEMONO_CREATOR)
	})
}
