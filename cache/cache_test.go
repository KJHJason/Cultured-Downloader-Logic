package cache

import (
	"context"
	"sync"
	"testing"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

func initTestData(t *testing.T, ctx context.Context) {
	err := InitCacheDb("")
	if err != nil {
		t.Fatalf("Failed to initialise cache db: %v", err)
	}
	CacheDb.ResetDb(ctx)

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

func initTestDataConcurrently(t *testing.T, ctx context.Context) {
	err := InitCacheDb("")
	if err != nil {
		t.Fatalf("Failed to initialise cache db: %v", err)
	}
	CacheDb.ResetDb(ctx)

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
		wg.Add(1); 
		defer wg.Done(); 
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initTestData(t, ctx)

	// Test getting the post key
	key := "https://fantia.jp/posts/123456"
	if !PostCacheExists(key, constants.FANTIA) {
		t.Errorf("Expected key to be in the cache")
	}

	// Test getting the post key
	val := GetTime(ParsePostKey(key, constants.FANTIA))
	if val.IsZero() {
		t.Errorf("Expected key to have a time value")
	}
}

func checkCacheValue(t *testing.T, expected []string, value ...*PostCache) {
	if len(value) != 2 {
		t.Errorf("Expected 2 values, got %d", len(value))
	}

	if value[0].Url != expected[0] {
		t.Errorf("Expected %s, got %s", expected[0], value[0])
	}
	if value[1].Url != expected[1] {
		t.Errorf("Expected %s, got %s", expected[1], value[1])
	}
}

func TestGetAllCacheForPlatform(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initTestData(t, ctx)

	// Test getting all cache for a platform
	cache := GetAllCacheForPlatform(ctx, constants.FANTIA)
	if len(cache) != 2 {
		t.Errorf("Expected 2 cache entries for Fantia")
	}
	checkCacheValue(
		t,
		[]string{"https://fantia.jp/posts/123456", "https://fantia.jp/posts/654321"},
		cache...,
	)

	cache = GetAllCacheForPlatform(ctx, constants.PIXIV)
	if len(cache) != 2 {
		t.Errorf("Expected 2 cache entries for Pixiv")
	}
	checkCacheValue(
		t,
		[]string{"https://www.pixiv.net/artworks/118849705", "https://www.pixiv.net/artworks/118849706"},
		cache...,
	)

	cache = GetAllCacheForPlatform(ctx, constants.PIXIV_FANBOX)
	if len(cache) != 2 {
		t.Errorf("Expected 2 cache entries for Pixiv Fanbox")
	}
	checkCacheValue(
		t,
		[]string{"https://creator.fanbox.cc/posts/7654321", "https://www.fanbox.cc/@creator/posts/1234567"},
		cache...,
	)

	cache = GetAllCacheForPlatform(ctx, constants.KEMONO)
	if len(cache) != 2 {
		t.Errorf("Expected 2 cache entries for Kemono")
	}
	checkCacheValue(
		t,
		[]string{"https://kemono.su/fanbox/user/1234567/post/1234567", "https://kemono.su/fanbox/user/1234567/post/7654321"},
		cache...,
	)
}

func TestConcurrency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initTestDataConcurrently(t, ctx)

	// Test getting the post key
	key := "https://fantia.jp/posts/123456"
	if !PostCacheExists(key, constants.FANTIA) {
		t.Errorf("Expected key to be in the cache")
	}

	// Test getting the post key
	val := GetTime(ParsePostKey(key, constants.FANTIA))
	if val.IsZero() {
		t.Errorf("Expected key to have a time value")
	}
}

func TestDeletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initTestData(t, ctx)

	// Test getting all cache for all platforms
	cache := GetAllCacheForAllPlatforms(ctx)
	if len(cache) != 8 {
		t.Errorf("Expected 8 cache entries")
	}

	// Test deleting all cache for all platforms
	err := DeletePostCacheForAllPlatforms(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test getting all cache for all platforms
	cache = GetAllCacheForAllPlatforms(ctx)
	if len(cache) != 0 {
		t.Errorf("Expected 0 cache entries")
	}
}

func TestListAllCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initTestData(t, ctx)

	// Test getting all cache
	var cache []*CacheKeyValue
	CacheDb.TraverseDb(ctx, func(key, val []byte) {
		cache = append(cache, &CacheKeyValue{Key: key, Val: val})
	})

	for _, c := range cache {
		var val string
		datetime := ParseBytesToDateTime(c.Val)
		if !datetime.IsZero() {
			val = datetime.String()
		} else {
			val = string(c.Val)
		}
		t.Logf("Key: %s, Val: %s", c.Key, val)
	}
}
