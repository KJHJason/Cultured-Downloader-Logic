package language

import (
	"testing"
)

func TestPrintAllValues(t *testing.T) {
	for key, lang := range translations {
		t.Logf("key: %v, EN: %v, JP: %v", key, lang.En, lang.Jp)
	}
}

func TestTranslationLogic(t *testing.T) {
	if val := Translate("home", "", EN); val != "Home" {
		t.Errorf("got: %v, expected: Home", val)
	}

	if val := Translate("home", "", JP); val != "ホーム" {
		t.Errorf("got: %v, expected: ホーム", val)
	}
}
