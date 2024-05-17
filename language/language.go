package language

import (
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/cache"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
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

func InitLangDb() {
	langDbPath := filepath.Join(iofuncs.APP_PATH, "language-db")

	var err error
	langDb, err = cache.NewDb(langDbPath)
	if err != nil {
		panic("failed to open language db: " + err.Error())
	}

	if DEBUG || needReseedDb() {
		if err := langDb.ResetDb(); err != nil {
			panic("failed to reset language db: " + err.Error())
		}

		initialiseDbData()
		logger.MainLogger.Info("Language database initialised")
	}
}

func Translate(key, lang string) string {
	fmtKey := strings.ToLower(key)
	fmtKey = strings.TrimSpace(fmtKey)
	if val := langDb.GetString(parseKey(fmtKey, lang)); val != "" {
		return val
	}
	return key
}

// IMPORTANT: PLEASE CLOSE THE DATABASE AFTER USE
func CloseDb() error {
	return langDb.Close()
}
