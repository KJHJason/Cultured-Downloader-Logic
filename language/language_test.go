package language

import (
	"testing"

	"github.com/KJHJason/Cultured-Downloader-Logic/database"
)

func TestPrintAllKV(t *testing.T) {
	defer database.CloseDb()

	InitLangDb(func(msg string) {
		t.Error(msg)
	})
	for _, kv := range database.AppDb.GetAllKeyValue(BUCKET) {
		t.Log(kv.GetKey(), kv.GetVal())
	}
}

func TestTranslationLogic(t *testing.T) {
	defer database.CloseDb()

	InitLangDb(func(msg string) {
		t.Error(msg)
	})
	t.Log(parseKey("home", EN))
	if val := Translate("home", "", EN); val != "Home" {
		t.Errorf("got: %v, expected: Home", val)
	}

	if val := Translate("home", "", JP); val != "ホーム" {
		t.Errorf("got: %v, expected: ホーム", val)
	}
}
