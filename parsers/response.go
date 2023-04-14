package parsers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

func logJsonResponse(body []byte) error {
	var prettyJson bytes.Buffer
	err := json.Indent(&prettyJson, body, "", "    ")
	if err != nil {
		err = fmt.Errorf(
			"error %d: failed to indent JSON response body due to %v",
			constants.JSON_ERROR,
			err,
		)
		logger.LogError(err, false, logger.ERROR)
		return err
	}

	filename := fmt.Sprintf("saved_%s.json", time.Now().Format("2006-01-02_15-04-05"))
	filePath := filepath.Join("json", filename)
	os.MkdirAll(filepath.Dir(filePath), 0666)
	err = os.WriteFile(filePath, prettyJson.Bytes(), 0666)
	if err != nil {
		err = fmt.Errorf(
			"error %d: failed to write JSON response body to file due to %v",
			constants.UNEXPECTED_ERROR,
			err,
		)
		logger.LogError(err, false, logger.ERROR)
		return err
	}
	return nil
}

// Read the response body and unmarshal it into a interface and returns it
func LoadJsonFromResponse(res *http.Response, format any) error {
	body, err := httpfuncs.ReadResBody(res)
	if err != nil {
		return err
	}

	// write to file if debug mode is on
	if constants.DEBUG_MODE {
		logJsonResponse(body)
	}

	if err = json.Unmarshal(body, &format); err != nil {
		return fmt.Errorf(
			"error %d: failed to unmarshal json response from %s due to %v\nBody: %s",
			constants.RESPONSE_ERROR,
			res.Request.URL.String(),
			err,
			string(body),
		)
	}
	return nil
}

func LoadJsonFromBytes(body []byte, format any) error {
	if err := json.Unmarshal(body, &format); err != nil {
		return fmt.Errorf(
			"error %d: failed to unmarshal json due to %v\nBody: %s",
			constants.JSON_ERROR,
			err,
			string(body),
		)
	}
	return nil
}
