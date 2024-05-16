package cache

import (
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

func PostCacheExists(key string) bool {
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
