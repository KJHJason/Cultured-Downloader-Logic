package cache

import (
	"context"
	"strings"
	"time"
)

// example of the Key-Value pairs in the database
// |------------|-----------|
// | key_suffix | unix_time |
// |------------|-----------|
// Where key_suffix can be <url>_<suffix> or <post_id>_<suffix>

const (
	POST           = "_post"
	UGOIRA         = "_ugoira"
	GDRIVE         = "_gdrive"
	KEMONO_CREATOR = "_kemono_creator"
)

func parseKey(key, category string) string {
	switch category {
	case POST:
		return key + POST
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

func separatePostKey(keyBytes []byte) (url string, platform string) {
	key := string(keyBytes)

	// key is in the format <url>_<platform>_post
	if !strings.HasSuffix(key, POST) {
		return "", ""
	}

	// remove the _post suffix
	key = key[:len(key)-len(POST)]

	// split the key into <url> and <platform>
	splitKey := strings.Split(key, "_")
	if len(splitKey) < 2 {
		// shouldn't happen but just in case
		return "", ""
	}

	url = strings.Join(splitKey[:len(splitKey)-1], "_")
	platform = splitKey[len(splitKey)-1]
	return url, platform
}

func PostCacheExists(key, platform string) bool {
	key += "_" + platform
	return len(Get(parseKey(key, POST))) > 0
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

func CachePost(key string) {
	SetTime(parseKey(key, POST), time.Now())
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

func GetAllCacheForPlatform(ctx context.Context, platform string) []*PostCache {
	keys := CacheDb.GetCacheKeyValue(ctx, func(key, _ []byte) bool {
		_, keyPlatform := separatePostKey(key)
		return keyPlatform == platform
	})

	caches := make([]*PostCache, 0, len(keys))
	for i, key := range keys {
		url, _ := separatePostKey(key.Key)
		caches[i] = &PostCache{
			Url:      url,
			Platform: platform,
			Datetime: ParseBytesToDateTime(key.Val),
			CacheKey: key.GetKey(),
		}
	}
	return caches
}

func DeletePostCacheForPlatform(ctx context.Context, platform string) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		_, keyPlatform := separatePostKey(key)
		return keyPlatform != platform
	})
}

func DeleteGdriveCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), GDRIVE)
	})
}

func DeleteUgoiraCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), UGOIRA)
	})
}

func DeleteKemonoCreatorCache(ctx context.Context) error {
	return CacheDb.ResetDbWithCond(ctx, func(key, _ []byte) bool {
		return strings.HasSuffix(string(key), KEMONO_CREATOR)
	})
}
