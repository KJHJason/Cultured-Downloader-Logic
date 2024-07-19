package language

import (
	_ "embed"
	"encoding/json"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

const (
	EN = constants.EN
	JP = constants.JP
)

var (
	//go:embed translations.json
	translationsJson []byte

	translations map[string]translationsValues
)

type translationsValues struct {
	En string `json:"en"`
	Jp string `json:"jp"`
}

func validateLang(lang string) string {
	if lang == JP {
		return JP
	}
	return EN
}

func init() {
	var translationsWithSections map[string]map[string]translationsValues
	err := json.Unmarshal(translationsJson, &translationsWithSections)
	if err != nil {
		panic(err)
	}
	translationsJson = nil

	translations = make(map[string]translationsValues)
	for _, sections := range translationsWithSections {
		for key, val := range sections {
			translations[key] = val
		}
	}
}
