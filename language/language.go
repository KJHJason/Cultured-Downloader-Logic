package language

import (
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/cache"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

var langDb *cache.DbWrapper

const (
	DEBUG           = true
	VERSION_KEY     = "version"
	CURRENT_VERSION = 0
)

func needReseedDb() bool {
	version := langDb.GetInt(VERSION_KEY)
	return DEBUG || CURRENT_VERSION > version
}

func init() {
	langDbPath := filepath.Join(iofuncs.APP_PATH, "language-db")

	var err error
	langDb, err = cache.NewDb(langDbPath)
	if err != nil {
		panic("failed to open language db: " + err.Error())
	}

	if DEBUG || needReseedDb() {
		initialiseDbData()
	}
}

func Translate(key, lang string) string {
	if val := langDb.GetString(parseKey(key, lang)); val != "" {
		return val
	}
	return key
}
