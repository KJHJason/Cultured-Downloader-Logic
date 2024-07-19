package language

import (
	"strings"
)

func Translate(key, fallback, lang string) string {
	fmtKey := strings.ToLower(key)
	fmtKey = strings.TrimSpace(fmtKey)
	lang = validateLang(lang)
	if val, ok := translations[fmtKey]; ok {
		switch lang {
		case JP:
			return val.Jp
		default:
			return val.En
		}
	}

	if fallback != "" {
		return fallback
	}
	return key
}
