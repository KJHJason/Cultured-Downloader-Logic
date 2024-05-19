package language

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

const (
	EN = constants.EN
	JP = constants.JP

	BUCKET          = "translations"
	DEBUG           = false
	VERSION_KEY     = "version"
	CURRENT_VERSION = 0
)

var (
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
	version := database.AppDb.GetInt(BUCKET, VERSION_KEY)
	return DEBUG || CURRENT_VERSION > version
}

func InitLangDb(panicHandler func(msg string)) {
	database.InitAppDb()
	if DEBUG || needReseedDb() {
		if err := database.AppDb.DeleteBucket(BUCKET); err != nil {
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

func handleInitErr(errMsg string, panicHandler func(msg string)) {
	if panicHandler != nil {
		panicHandler(errMsg)
	}
	logger.MainLogger.Fatal(errMsg)
}

func initialiseDbData(panicHandler func(msg string)) {
	tx, err := database.AppDb.Db.Begin(true)
	if err != nil {
		handleInitErr(
			fmt.Sprintf("failed to start transaction: %v", err),
			panicHandler,
		)
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(BUCKET))
	if err != nil {
		handleInitErr(
			fmt.Sprintf("failed to create bucket: %v", err),
			panicHandler,
		)
	}

	currentVer := database.ParseInt(CURRENT_VERSION)
	if err := b.Put([]byte(VERSION_KEY), currentVer); err != nil {
		handleInitErr(
			fmt.Sprintf("failed to set version: %v", err),
			panicHandler,
		)
	}

	var translations map[string]map[string]translations
	err = json.Unmarshal(translationsJson, &translations)
	if err != nil {
		handleInitErr(
			fmt.Sprintf("failed to unmarshal translations: %v", err),
			panicHandler,
		)
	}

	for _, section := range translations {
		for key, translation := range section {
			key = strings.ToLower(key)
			key = strings.TrimSpace(key)

			enKey := parseKey(key, EN)
			if err := b.Put([]byte(enKey), []byte(translation.En)); err != nil {
				handleInitErr(
					fmt.Sprintf("failed to set key (EN): %v", err),
					panicHandler,
				)
			}

			jpKey := parseKey(key, JP)
			if err := b.Put([]byte(jpKey), []byte(translation.Jp)); err != nil {
				handleInitErr(
					fmt.Sprintf("failed to set key (JP): %v", err),
					panicHandler,
				)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		handleInitErr(
			fmt.Sprintf("failed to commit translations to db: %v", err),
			panicHandler,
		)
	}
}
