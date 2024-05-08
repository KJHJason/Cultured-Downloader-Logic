package extractor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/mholt/archiver/v4"
	"github.com/jbenet/go-context/io"
)

type archiveExtractor struct {
	reader     io.Reader
	readCloser io.ReadCloser
	ex         archiver.Extractor
}

func extractFileLogic(ctx context.Context, src, dest string, extractor *archiveExtractor) error {
	handler := func(ctx context.Context, file archiver.File) error {
		extractedFilePath := filepath.Join(dest, file.NameInArchive)
		os.MkdirAll(filepath.Dir(extractedFilePath), constants.DEFAULT_PERMS)

		af, err := file.Open()
		if err != nil {
			return err
		}
		defer af.Close()

		out, err := os.OpenFile(
			extractedFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, ctxio.NewReader(ctx, af))
		if err != nil {
			return err
		}
		return nil
	}

	var input io.Reader
	if extractor.readCloser != nil {
		input = extractor.readCloser
	} else {
		input = extractor.reader
	}

	err := extractor.ex.Extract(ctx, input, nil, handler)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			// delete all the files that were extracted
			if osErr := os.RemoveAll(dest); osErr != nil {
				logger.LogError(osErr, false, logger.ERROR)
			}
			return err
		}
		return fmt.Errorf(
			"error %d: unable to extract zip file %s, more info => %w",
			errs.OS_ERROR,
			src,
			err,
		)
	}
	return nil
}

func getExtractor(f *os.File, src string) (*archiveExtractor, error) {
	filename := filepath.Base(src)
	format, archiveReader, err := archiver.Identify(filename, f)
	if errors.Is(err, archiver.ErrNoMatch) {
		return nil, fmt.Errorf(
			"error %d: %s is not a valid zip file",
			errs.OS_ERROR,
			src,
		)
	} else if err != nil {
		return nil, err
	}

	var rc io.ReadCloser
	if decom, ok := format.(archiver.Decompressor); ok {
		rc, err = decom.OpenReader(archiveReader)
		if err != nil {
			return nil, err
		}
	}

	ex, ok := format.(archiver.Extractor)
	if !ok {
		return nil, fmt.Errorf(
			"error %d: unable to extract zip file %s, more info => %w",
			errs.UNEXPECTED_ERROR,
			src,
			err,
		)
	}
	return &archiveExtractor{
		reader:     archiveReader,
		readCloser: rc,
		ex:         ex,
	}, nil
}

func getErrIfNotIgnored(src string, ignoreIfMissing bool) error {
	if ignoreIfMissing {
		return nil
	} 
	return fmt.Errorf(
		"error %d: %s does not exist",
		errs.OS_ERROR,
		src,
	)
}

// Extract all files from the given archive file to the given destination
//
// Code based on https://stackoverflow.com/a/24792688/2737403
func ExtractFiles(ctx context.Context, src, dest string, ignoreIfMissing bool) error {
	if !iofuncs.PathExists(src) {
		return getErrIfNotIgnored(src, ignoreIfMissing)
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf(
			"error %d: unable to open zip file %s",
			errs.OS_ERROR,
			src,
		)
	}
	defer f.Close()

	extractor, err := getExtractor(f, src)
	if err != nil {
		return err
	}

	if extractor.readCloser != nil {
		defer extractor.readCloser.Close()
	}
	return extractFileLogic(
		ctx, 
		src, 
		dest,
		extractor,
	)
}
