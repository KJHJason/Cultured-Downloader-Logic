package database

import (
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

var (
	AppDb       *DbWrapper
	APP_DB_PATH = filepath.Join(iofuncs.APP_PATH, "app.db")
)

func InitAppDb() error {
	// if the cache db is already initialised, return
	if AppDb != nil {
		return nil
	}

	var err error
	AppDb, err = NewDb(APP_DB_PATH)
	if err != nil {
		return err
	}
	return nil
}

func CloseDb() error {
	if AppDb == nil {
		return nil
	}

	if err := AppDb.Close(); err != nil {
		return err
	}
	AppDb = nil
	return nil
}
