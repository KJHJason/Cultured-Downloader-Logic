package cf

import (
	"encoding/json"
	"fmt"
	"os"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

type Cookie struct {
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Domain       string  `json:"domain"`
	Path         string  `json:"path"`
	Expires      float64 `json:"expires"`
	Size         int     `json:"size"`
	HttpOnly     bool    `json:"httpOnly"`
	Secure       bool    `json:"secure"`
	Session      bool    `json:"session"`
	Priority     string  `json:"priority"`
	SameParty    bool    `json:"sameParty"`
	SourceScheme string  `json:"sourceScheme"`
	SourcePort   int     `json:"sourcePort"`
}

type Cookies []*Cookie

func parseCookies(filePath string) (Cookies, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to read cookies file %s => %w",
			cdlerrors.OS_ERROR,
			filePath,
			err,
		)
	}

	var cookies Cookies
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to unmarshal cookies data => %w",
			cdlerrors.JSON_ERROR,
			err,
		)
	}
	return cookies, nil
}
