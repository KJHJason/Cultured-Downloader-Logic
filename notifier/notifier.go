package notifier

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/gen2brain/beeep"
	"fyne.io/fyne/v2"
)

var (
	//go:embed icon.png
	iconImg []byte
	iconPath = filepath.Join(iofuncs.APP_PATH, "icon.png")
)

const CLI_TITLE = "Cultured Downloader CLI"

func writeIcon() error {
	defer func() {
		if iconImg != nil {
			iconImg = nil
		}
	}()

	if iofuncs.PathExists(iconPath) {
		return nil
	}

	f, err := os.Create(iconPath)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, bytes.NewReader(iconImg)); err != nil {
		return err
	}
	return nil
}

func fyneAlert(title, message string, app fyne.App) {
	if app == nil {
		return
	}

	app.SendNotification(
		fyne.NewNotification(title, message),
	)
}

// Alert shows a notification on the user's system with the given title and message.
func Alert(title, message string, app fyne.App) error {	
	if title != CLI_TITLE {
		fyneAlert(title, message, app)
		return nil
	}

	if err := writeIcon(); err != nil {
		return fmt.Errorf(
			"error %d: unable to write notification icon => %v", 
			constants.UNEXPECTED_ERROR,
			err,
		)
	}

	if err := beeep.Alert(title, message, iconPath); err != nil {
		return fmt.Errorf(
			"error %d: unable to show notification => %v", 
			constants.UNEXPECTED_ERROR,
			err,
		)
	}

	return nil
}

// AlertWithoutErr is the same as Alert but 
// if an error occurs, it will log it instead of returning it.
func AlertWithoutErr(title, message string, app fyne.App) {
	if err := Alert(title, message, app); err != nil {
		logger.LogError(err, false, logger.ERROR)
	}
}
