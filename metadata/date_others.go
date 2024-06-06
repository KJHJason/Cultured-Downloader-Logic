//go:build linux || darwin
// +build linux darwin

package metadata

import (
	"context"
	"os"
	"time"
)

func ChangeFilePathCreationDate(ctx context.Context, filePath string, creationDate time.Time) error {
	metadata := Metadata{CreationDate: creationDate}
	return SetExifDataToImage(ctx, filePath, metadata)
}

func ChangeFileCreationDate(ctx context.Context, file *os.File, creationDate time.Time) error {
	panic("not implemented for this OS")
}
