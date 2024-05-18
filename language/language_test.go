package language

import (
	"context"
	"testing"

	"github.com/cockroachdb/pebble"
)

func TestPrintAllKV(t *testing.T) {
	defer CloseDb()

	InitLangDb(context.Background(), nil)
	iter, err := langDb.Db.NewIter(&pebble.IterOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		t.Logf("%s: %s", iter.Key(), iter.Value())
	}
}

func TestTranslationLogic(t *testing.T) {
	defer CloseDb()

	if Translate("home", EN, "") != "Home" {
		t.Error("expected: Home")
	}

	if Translate("home", JP, "") != "ホーム" {
		t.Error("expected: ホーム")
	}
}
