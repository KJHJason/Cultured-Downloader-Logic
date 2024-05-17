package language

import (
	"strings"
)

func Translate(key, fallback, lang string) string {
	fmtKey := strings.ToLower(key)
	fmtKey = strings.TrimSpace(fmtKey)
	if val := langDb.GetString(parseKey(fmtKey, lang)); val != "" {
		return val
	}

	if fallback != "" {
		return fallback
	}
	return key
}

// IMPORTANT: PLEASE CLOSE THE DATABASE AFTER USE
func CloseDb() error {
	return langDb.Close()
}
