package language

import (
	"github.com/cockroachdb/pebble"
)

const (
	EN = "en"
	JP = "ja"
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

func addData(key, value, lang string) {
	key = parseKey(key, lang)
	if err := langDb.Db.Set([]byte(key), []byte(value), pebble.Sync); err != nil {
		panic("failed to set key: " + err.Error())
	}
}

func initialiseDbData() {
	batch := langDb.Db.NewBatch()
	addData("test", "test", EN)

	initForDownloadQueue()
	initForGeneral()
	initForHomePage()
	initForProgramInfo()
	initForPagination()

	if err := langDb.SetBatch(batch); err != nil {
		panic("failed to apply batch: " + err.Error())
	}
}
