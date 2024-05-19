package language

import (
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/database"
)

func Translate(key, fallback, lang string) string {
	fmtKey := strings.ToLower(key)
	fmtKey = strings.TrimSpace(fmtKey)
	fmtKey = parseKey(fmtKey, lang)
	if val := database.AppDb.GetString(BUCKET, fmtKey); val != "" {
		return val
	}

	if fallback != "" {
		return fallback
	}
	return key
}
