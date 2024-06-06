//go:build windows
// +build windows

package metadata

import (
	"context"
	"os"
	"syscall"
	"time"
)

func ChangeFilePathCreationDate(ctx context.Context, filePath string, creationDate time.Time) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	return ChangeFileCreationDate(ctx, file, creationDate)
}

func ChangeFileCreationDate(ctx context.Context, file *os.File, creationDate time.Time) error {
	fh := syscall.Handle(file.Fd())

	creationTime := syscall.NsecToFiletime(creationDate.UnixNano())

	err := syscall.SetFileTime(fh, &creationTime, nil, nil)
	if err != nil {
		return err
	}
	return nil
}
