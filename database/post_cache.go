package database

import (
	"fmt"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

// example of the Key-Value pairs in the database
// |------------|-----------|
// | prefix_key | unix_time |
// |------------|-----------|
// Where key_suffix can be <prefix>|<platform>|post_url_or_id

const (
	POST_DELIM            = "|"
	POST_BUCKET           = "post_cache"
	GDRIVE_BUCKET         = "gdrive_cache"
	UGOIRA_BUCKET         = "ugoira_cache"
	KEMONO_CREATOR_BUCKET = "kemono_creator_cache"
)

func ParsePostKey(url, platform string) string {
	return platform + POST_DELIM + url
}

func SeparatePostKey(keyBytes []byte) (url string, platform string) {
	// key should be in the format of <platform>|<url>
	key := string(keyBytes)

	// split the key into <url> and <platform>
	splitKey := strings.Split(key, POST_DELIM)
	splitLen := len(splitKey)
	if splitLen < 2 {
		// shouldn't happen but just in case
		return "", ""
	}

	platform = splitKey[0]
	if splitLen == 2 {
		return splitKey[1], platform
	}

	// shouldn't have an edge case where the url
	// contains the platform suffix but just in case
	url = strings.Join(splitKey[1:], POST_DELIM)
	return url, platform
}

func PostCacheExists(key, platform string) bool {
	return len(AppDb.GetString(POST_BUCKET, ParsePostKey(key, platform))) > 0
}

func getPostCache(key, platform string) time.Time {
	return AppDb.GetTime(POST_BUCKET, ParsePostKey(key, platform))
}

func GDriveCacheExists(key string) bool {
	return len(AppDb.GetString(GDRIVE_BUCKET, key)) > 0
}

func UgoiraCacheExists(key string) bool {
	return len(AppDb.GetString(UGOIRA_BUCKET, key)) > 0
}

func GetKemonoCreatorCache(key string) string {
	return AppDb.GetString(KEMONO_CREATOR_BUCKET, key)
}

// Note: the setter functions below do not handle errors and will continue as usual

func CachePost(parsedKey string) {
	AppDb.SetTime(POST_BUCKET, parsedKey, time.Now())
}

func batchCacheLogic(tx *bolt.Tx, bucketName string, key string) error {
	b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	return b.Put([]byte(key), ParseDateTimeToBytes(time.Now()))
}

func CachePostViaBatch(parsedKey string) {
	AppDb.Db.Batch(func(tx *bolt.Tx) error {
		return batchCacheLogic(tx, POST_BUCKET, parsedKey)
	})
}

func CacheGDrive(key string) {
	AppDb.Db.Batch(func(tx *bolt.Tx) error {
		return batchCacheLogic(tx, GDRIVE_BUCKET, key)
	})
}

func CacheUgoira(key string) {
	AppDb.Db.Batch(func(tx *bolt.Tx) error {
		return batchCacheLogic(tx, UGOIRA_BUCKET, key)
	})
}

func CacheKemonoCreatorName(key, creatorName string) {
	AppDb.Db.Batch(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(KEMONO_CREATOR_BUCKET))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), []byte(creatorName))
	})
}

type PostCache struct {
	Url      string    `json:"Url"`
	Platform string    `json:"Platform"`
	Datetime time.Time `json:"Datetime"`
	CacheKey string    `json:"CacheKey"`
	Bucket   string    `json:"Bucket"`
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

func sortPostCacheByDatetime(caches []*PostCache) {
	sort.Slice(caches, func(i, j int) bool {
		return caches[i].Datetime.After(caches[j].Datetime)
	})
}

func GetAllCacheForPlatform(platforms ...string) []*PostCache {
	var caches []*KeyValue
	for _, platform := range platforms {
		caches = append(caches, AppDb.GetKeyValueOnPrefix(POST_BUCKET, platform)...)
	}

	if len(caches) == 0 {
		return make([]*PostCache, 0)
	}

	postCache := make([]*PostCache, 0, len(caches))
	for _, cache := range caches {
		url, platform := SeparatePostKey(cache.Key)
		postCache = append(postCache, &PostCache{
			Url:      url,
			Platform: platform,
			Datetime: ParseBytesToDateTime(cache.Val),
			CacheKey: string(cache.Key),
			Bucket:   POST_BUCKET,
		})
	}
	sortPostCacheByDatetime(postCache)
	return postCache
}

func GetAllCacheForAllPlatforms() []*PostCache {
	postCache := AppDb.GetAllKeyValue(POST_BUCKET)
	if len(postCache) == 0 {
		return make([]*PostCache, 0)
	}

	allCache := make([]*PostCache, 0, len(postCache))
	for _, cache := range postCache {
		url, platform := SeparatePostKey(cache.Key)
		allCache = append(allCache, &PostCache{
			Url:      url,
			Platform: platform,
			Datetime: ParseBytesToDateTime(cache.Val),
			CacheKey: string(cache.Key),
			Bucket:   POST_BUCKET,
		})
	}
	sortPostCacheByDatetime(allCache)
	return allCache
}

func DeletePostCacheForAllPlatforms() error {
	return AppDb.DeleteBucket(POST_BUCKET)
}

func GetAllGdriveCache() []*KeyValue {
	return AppDb.GetAllKeyValue(GDRIVE_BUCKET)
}

func DeleteAllGdriveCache() error {
	return AppDb.DeleteBucket(GDRIVE_BUCKET)
}

func GetAllUgoiraCache() []*KeyValue {
	return AppDb.GetAllKeyValue(UGOIRA_BUCKET)
}

func DeleteAllUgoiraCache() error {
	return AppDb.DeleteBucket(UGOIRA_BUCKET)
}

func GetAllKemonoCreatorCache() []*KeyValue {
	return AppDb.GetAllKeyValue(KEMONO_CREATOR_BUCKET)
}

func DeleteAllKemonoCreatorCache() error {
	return AppDb.DeleteBucket(KEMONO_CREATOR_BUCKET)
}
