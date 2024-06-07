package metadata

import (
	"context"
)

func ChangeFilePathCreationDate(ctx context.Context, filePath string, metadata Metadata) error {
	return SetExifDataToImage(ctx, filePath, metadata)
}
