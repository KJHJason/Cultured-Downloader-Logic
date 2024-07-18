package cf

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

const VERSION = "v0.1.0"

func getTestArgs() []string {
	chromePath, err := utils.GetChromeExecPath()
	if err != nil {
		// chrome exec path check should have been done
		// at the start of the program. Hence, the panic here.
		panic(err)
	}

	return []string{
		"--test-connection",
		"--headless=true",
		"--browser-path", chromePath,
		"--log-path", logger.CdlCfLogFilePath,
	}
}
