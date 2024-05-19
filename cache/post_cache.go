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
	return len(Get(key + GDRIVE)) > 0
}

func UgoiraCacheExists(key string) bool {
	return len(Get(key + UGOIRA)) > 0
}

func GetKemonoCreatorCache(key string) string {
	return GetString(key + KEMONO_CREATOR)
}

// Note: the setter functions below do not handle errors and will continue as usual

func CachePost(parsedKey string) {
	SetTime(parsedKey, time.Now())
}

func CacheGDrive(key string) {
	SetTime(key + GDRIVE, time.Now())
}

func CacheUgoira(key string) {
	SetTime(key + UGOIRA, time.Now())
}

func CacheKemonoCreatorName(key, creatorName string) {
	SetString(key + KEMONO_CREATOR, creatorName)
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
	var caches []*PostCache
	CacheDb.TraverseDb(ctx, func(key, value []byte) {
		url, keyPlatform := SeparatePostKey(key)
		if url == "" || keyPlatform == "" {
			return
		}
		for _, p := range platforms {
			if keyPlatform == p {
				caches = append(caches, &PostCache{
					Url:      url,
					Platform: GetReadableSiteStrSafely(keyPlatform),
					Datetime: ParseBytesToDateTime(value),
					CacheKey: string(key),
				})
				return
			}
		}
	})
	if len(caches) == 0 {
		return make([]*PostCache, 0)
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
	var cacheKeys []*CacheKeyValue
	CacheDb.TraverseDb(ctx, func(key, val []byte) {
		if strings.HasSuffix(string(key), GDRIVE) {
			cacheKeys = append(cacheKeys, &CacheKeyValue{Key: key, Val: val})
		}
	})
	return cacheKeys
}

func DeleteAllGdriveCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), GDRIVE)
	})
}

func GetAllUgoiraCache(ctx context.Context) []*CacheKeyValue {
	var cacheKeys []*CacheKeyValue
	CacheDb.TraverseDb(ctx, func(key, val []byte) {
		if strings.HasSuffix(string(key), UGOIRA) {
			cacheKeys = append(cacheKeys, &CacheKeyValue{Key: key, Val: val})
		}
	})
	return cacheKeys
}

func DeleteAllUgoiraCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), UGOIRA)
	})
}

func GetAllKemonoCreatorCache(ctx context.Context) []*CacheKeyValue {
	var cacheKeys []*CacheKeyValue
	CacheDb.TraverseDb(ctx, func(key, val []byte) {
		if strings.HasSuffix(string(key), KEMONO_CREATOR) {
			cacheKeys = append(cacheKeys, &CacheKeyValue{Key: key, Val: val})
		}
	})
	return cacheKeys
}

func DeleteAllKemonoCreatorCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), KEMONO_CREATOR)
	})
}
