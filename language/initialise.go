package language

import (
	"github.com/cockroachdb/pebble"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

const (
	EN = constants.EN
	JP = constants.JP
)

// example of the Key-Value pairs in the database
// |---------|-----------------|
// | text_en | text_in_english |
// |---------|-----------------|
// | text_jp | text_in_japanese|
// |---------|-----------------|

func parseKey(key, lang string) string {
	return key + "_" + lang
}

type dataInitWrapper struct {
	batch *pebble.Batch
}

func (d *dataInitWrapper) addData(key, value, lang string) {
	key = parseKey(key, lang)
	if err := d.batch.Set([]byte(key), []byte(value), pebble.Sync); err != nil {
		panic("failed to set key: " + err.Error())
	}
}

func initialiseDbData() {
	db := &dataInitWrapper{ batch: langDb.Db.NewBatch() }
	db.addData("test", "test", EN)

	initForDownloadQueue(db)
	initForGeneral(db)
	initForHomePage(db)
	initForProgramInfo(db)
	initForPagination(db)

	if err := langDb.SetBatch(db.batch); err != nil {
		panic("failed to apply batch: " + err.Error())
	}
}
