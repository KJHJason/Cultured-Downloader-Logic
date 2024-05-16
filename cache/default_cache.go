package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	cp "github.com/otiai10/copy"
)

var (
	DEFAULT_PATH            = filepath.Join(iofuncs.APP_PATH, "cache")
	CacheDb      *DbWrapper = nil
)

func InitCache(path string) error {
	if path == "" {
		path = DEFAULT_PATH
	}

	db, err := NewDb(path)
	if err != nil {
		return err
	}
	CacheDb = db
	return nil
}

func MoveDb(oldPath, newPath string) error {
	if CacheDb != nil {
		if err := CacheDb.Close(); err != nil {
			return err
		}
	}

	// move the directory containing the cache
	err := cp.Copy(oldPath, newPath)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(oldPath); err != nil {
		logger.MainLogger.Errorf("Failed to remove old cache directory: %s", err)
	}
	return InitCache(newPath)
}

func Get(key string) []byte {
	return CacheDb.Get(key)
}

func GetJson(key string, v any) error {
	return CacheDb.GetJson(key, v)
}

func GetString(key string) string {
	return string(CacheDb.Get(key))
}

func GetInt64(key string) int64 {
	return CacheDb.GetInt64(key)
}

func GetInt(key string) int {
	return CacheDb.GetInt(key)
}

func GetTime(key string) time.Time {
	return CacheDb.GetTime(key)
}

func Set(key string, value []byte) error {
	return CacheDb.Set(key, value)
}

func SetJson(key string, v any) error {
	return CacheDb.SetJson(key, v)
}

func SetString(key string, value string) error {
	return CacheDb.SetString(key, value)
}

func SetInt64(key string, value int64) error {
	return CacheDb.SetInt64(key, value)
}

func SetInt(key string, value int) error {
	return CacheDb.SetInt(key, value)
}

func SetTime(key string, value time.Time) error {
	return CacheDb.SetTime(key, value)
}

func Delete(key string) error {
	return CacheDb.Delete(key)
}
