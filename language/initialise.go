package language

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/cache"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/cockroachdb/pebble"
)

const (
	EN = constants.EN
	JP = constants.JP

	DEBUG           = false
	VERSION_KEY     = "version"
	CURRENT_VERSION = 0
)

var (
	langDb *cache.DbWrapper

	//go:embed translations.json
	translationsJson []byte
)

type translations struct {
	En string `json:"en"`
	Jp string `json:"jp"`
}

// example of the Key-Value pairs in the database
// |---------|-----------------|
// | text_en | text_in_english |
// |---------|-----------------|
// | text_jp | text_in_japanese|
// |---------|-----------------|

func needReseedDb() bool {
	version := langDb.GetInt(VERSION_KEY)
	return DEBUG || CURRENT_VERSION > version
}

func InitLangDb(ctx context.Context, panicHandler func(msg string)) {
	langDbPath := filepath.Join(iofuncs.APP_PATH, "language-db")

	var err error
	langDb, err = cache.NewDb(langDbPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to open language db: %v", err)
		if panicHandler != nil {
			panicHandler(errMsg)
		}
		logger.MainLogger.Fatal(errMsg)
	}

	if DEBUG || needReseedDb() {
		if err := langDb.ResetDb(ctx); err != nil {
			errMsg := fmt.Sprintf("failed to reset language db: %v", err)
			if panicHandler != nil {
				panicHandler(errMsg)
			}
			logger.MainLogger.Fatal(errMsg)
		}

		initialiseDbData(panicHandler)
		logger.MainLogger.Info("Language database initialised")
	}
	translationsJson = nil
}

func parseKey(key, lang string) string {
	return key + "_" + lang
}

type dataInitWrapper struct {
	batch *pebble.Batch
}

func (d *dataInitWrapper) addTranslations(key string, translations translations) {
	key = strings.ToLower(key)
	key = strings.TrimSpace(key)

	enKey := parseKey(key, EN)
	if err := d.batch.Set([]byte(enKey), []byte(translations.En), pebble.Sync); err != nil {
		logger.MainLogger.Fatalf("failed to set key (EN): %v", err)
	}

	jpKey := parseKey(key, JP)
	if err := d.batch.Set([]byte(jpKey), []byte(translations.Jp), pebble.Sync); err != nil {
		logger.MainLogger.Fatalf("failed to set key (JP): %v", err)
	}
}

func initialiseDbData(panicHandler func(msg string)) {
	db := &dataInitWrapper{batch: langDb.Db.NewBatch()}

	currentVer := cache.ParseInt(CURRENT_VERSION)
	db.batch.Set([]byte(VERSION_KEY), currentVer, pebble.Sync)

	var translations map[string]map[string]translations
	err := json.Unmarshal(translationsJson, &translations)
	if err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal translations: %v", err)
		if panicHandler != nil {
			panicHandler(errMsg)
		}
		logger.MainLogger.Fatal(errMsg)
	}

	for _, section := range translations {
		for key, translation := range section {
			db.addTranslations(key, translation)
		}
	}

	if err := langDb.SetBatch(db.batch); err != nil {
		errMsg := fmt.Sprintf("failed to apply translations batch: %v", err)
		if panicHandler != nil {
			panicHandler(errMsg)
		}
		logger.MainLogger.Fatal(errMsg)
	}
}
