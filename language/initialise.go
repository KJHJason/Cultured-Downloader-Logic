package language

import (
	"context"
	_ "embed"
	"encoding/json"
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

func InitLangDb(ctx context.Context) {
	langDbPath := filepath.Join(iofuncs.APP_PATH, "language-db")

	var err error
	langDb, err = cache.NewDb(langDbPath)
	if err != nil {
		panic("failed to open language db: " + err.Error())
	}

	if DEBUG || needReseedDb() {
		if err := langDb.ResetDb(ctx); err != nil {
			panic("failed to reset language db: " + err.Error())
		}

		initialiseDbData()
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
		panic("failed to set key: " + err.Error())
	}

	jpKey := parseKey(key, JP)
	if err := d.batch.Set([]byte(jpKey), []byte(translations.Jp), pebble.Sync); err != nil {
		panic("failed to set key: " + err.Error())
	}
}

func initialiseDbData() {
	db := &dataInitWrapper{batch: langDb.Db.NewBatch()}

	currentVer := cache.ParseInt(CURRENT_VERSION)
	db.batch.Set([]byte(VERSION_KEY), currentVer, pebble.Sync)

	var translations map[string]map[string]translations
	err := json.Unmarshal(translationsJson, &translations)
	if err != nil {
		panic("failed to unmarshal translations: " + err.Error())
	}

	for _, section := range translations {
		for key, translation := range section {
			db.addTranslations(key, translation)
		}
	}

	if err := langDb.SetBatch(db.batch); err != nil {
		panic("failed to apply batch: " + err.Error())
	}
}
