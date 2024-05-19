package database

import (
	"sync"
	"testing"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

func resetBuckets() {
	AppDb.DeleteBucket(POST_BUCKET)
	AppDb.DeleteBucket(GDRIVE_BUCKET)
	AppDb.DeleteBucket(UGOIRA_BUCKET)
	AppDb.DeleteBucket(KEMONO_CREATOR_BUCKET)
}

func initTestData(t *testing.T) {
	err := InitAppDb()
	if err != nil {
		t.Fatalf("Failed to initialise cache db: %v", err)
	}
	resetBuckets()

	// Add some data to the cache
	CachePost(ParsePostKey("https://fantia.jp/posts/123456", constants.FANTIA))
	CachePost(ParsePostKey("https://fantia.jp/posts/654321", constants.FANTIA))
	CachePost(ParsePostKey("https://www.pixiv.net/artworks/118849705", constants.PIXIV))
	CachePost(ParsePostKey("https://www.pixiv.net/artworks/118849706", constants.PIXIV))
	CachePost(ParsePostKey("https://www.fanbox.cc/@creator/posts/1234567", constants.PIXIV_FANBOX))
	CachePost(ParsePostKey("https://creator.fanbox.cc/posts/7654321", constants.PIXIV_FANBOX))
	CachePost(ParsePostKey("https://kemono.su/fanbox/user/1234567/post/1234567", constants.KEMONO))
	CachePost(ParsePostKey("https://kemono.su/fanbox/user/1234567/post/7654321", constants.KEMONO))

	CacheGDrive("https://drive.google.com/file/d/<file_id>/view?usp=drive_link")
	CacheUgoira("https://www.pixiv.net/artworks/118849705")
	CacheKemonoCreatorName("https://kemono.su/fanbox/user/1234567", "Kemono Creator")
}

func initTestDataConcurrently(t *testing.T) {
	err := InitAppDb()
	if err != nil {
		t.Fatalf("Failed to initialise cache db: %v", err)
	}
	resetBuckets()

	// Add some data to the cache
	wg := sync.WaitGroup{}
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://fantia.jp/posts/123456", constants.FANTIA))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://fantia.jp/posts/654321", constants.FANTIA))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://www.pixiv.net/artworks/118849705", constants.PIXIV))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://www.pixiv.net/artworks/118849706", constants.PIXIV))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://www.fanbox.cc/@creator/posts/1234567", constants.PIXIV_FANBOX))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://creator.fanbox.cc/posts/7654321", constants.PIXIV_FANBOX))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://kemono.su/fanbox/user/1234567/post/1234567", constants.KEMONO))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CachePost(ParsePostKey("https://kemono.su/fanbox/user/1234567/post/7654321", constants.KEMONO))
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CacheGDrive("https://drive.google.com/file/d/<file_id>/view?usp=drive_link")
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CacheUgoira("https://www.pixiv.net/artworks/118849705")
	}()
	func() {
		wg.Add(1)
		defer wg.Done()
		go CacheKemonoCreatorName("https://kemono.su/fanbox/user/1234567", "Kemono Creator")
	}()
	wg.Wait()
}

func TestCachePost(t *testing.T) {
	initTestData(t)

	// Test getting the post key
	key := "https://fantia.jp/posts/123456"
	if !PostCacheExists(key, constants.FANTIA) {
		t.Errorf("Expected key to be in the cache")
	}

	// Test getting the post key
	val := getPostCache(key, constants.FANTIA)
	if val.IsZero() {
		t.Errorf("Expected key to have a time value")
	}
}

func checkCacheValue(t *testing.T, expected []string, value ...*PostCache) {
	if len(value) != 2 {
		t.Errorf("Expected 2 values, got %d", len(value))
	}

	count := 0
	for _, v := range value {
		for _, e := range expected {
			if v.Url == e {
				count++
				break
			}
		}
	}
	if count != 2 {
		t.Errorf("Expected 2 values to match the expected values")
	}
}

func TestGetAllCacheForPlatform(t *testing.T) {
	initTestData(t)

	// Test getting all cache for a platform
	cache := GetAllCacheForPlatform(constants.FANTIA)
	if len(cache) != 2 {
		t.Fatalf("Expected 2 cache entries for Fantia, got %d", len(cache))
	}
	checkCacheValue(
		t,
		[]string{"https://fantia.jp/posts/123456", "https://fantia.jp/posts/654321"},
		cache...,
	)

	cache = GetAllCacheForPlatform(constants.PIXIV)
	if len(cache) != 2 {
		t.Fatalf("Expected 2 cache entries for Pixiv, got %d", len(cache))
	}
	checkCacheValue(
		t,
		[]string{"https://www.pixiv.net/artworks/118849705", "https://www.pixiv.net/artworks/118849706"},
		cache...,
	)

	cache = GetAllCacheForPlatform(constants.PIXIV_FANBOX)
	if len(cache) != 2 {
		t.Fatalf("Expected 2 cache entries for Pixiv Fanbox, got %d", len(cache))
	}
	checkCacheValue(
		t,
		[]string{"https://creator.fanbox.cc/posts/7654321", "https://www.fanbox.cc/@creator/posts/1234567"},
		cache...,
	)

	cache = GetAllCacheForPlatform(constants.KEMONO)
	if len(cache) != 2 {
		t.Fatalf("Expected 2 cache entries for Kemono, got %d", len(cache))
	}
	checkCacheValue(
		t,
		[]string{"https://kemono.su/fanbox/user/1234567/post/1234567", "https://kemono.su/fanbox/user/1234567/post/7654321"},
		cache...,
	)
}

func TestConcurrency(t *testing.T) {
	initTestDataConcurrently(t)

	// Test getting the post key
	key := "https://fantia.jp/posts/123456"
	if !PostCacheExists(key, constants.FANTIA) {
		t.Errorf("Expected key to be in the cache")
	}

	// Test getting the post key
	val := getPostCache(key, constants.FANTIA)
	if val.IsZero() {
		t.Errorf("Expected key to have a time value")
	}
}

func TestDeletion(t *testing.T) {
	initTestData(t)

	// Test getting all cache for all platforms
	cache := GetAllCacheForAllPlatforms()
	if len(cache) != 8 {
		t.Errorf("Expected 8 cache entries")
	}

	// Test deleting all cache for all platforms
	err := DeletePostCacheForAllPlatforms()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test getting all cache for all platforms
	cache = GetAllCacheForAllPlatforms()
	if len(cache) != 0 {
		t.Errorf("Expected 0 cache entries")
	}
}
