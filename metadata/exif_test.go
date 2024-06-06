package metadata

import (
	"context"
	"io"
	"os"
	"testing"
	"time"
)

func getFileExt() string {
	return "png"
}

func getTestImgCopyPath() string {
	return "testc." + getFileExt()
}

func getTestImgPath() string {
	return "test." + getFileExt()
}

func TestWriteExifDataToImage(t *testing.T) {
	testImgPath := getTestImgPath()
	// create a copy of the test image
	src, err := os.Open(testImgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	dstFile := getTestImgCopyPath()
	dst, err := os.Create(dstFile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		t.Fatal(err)
	}
	dst.Close()

	err = ChangeFilePathCreationDate(context.Background(), dstFile, time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
}

func TestSizeDiff(t *testing.T) {
	TestWriteExifDataToImage(t)

	oriFile, err := os.Stat(getTestImgPath())
	if err != nil {
		t.Fatal(err)
	}
	newFile, err := os.Stat(getTestImgCopyPath())
	if err != nil {
		t.Fatal(err)
	}

	oriSize := oriFile.Size()
	newSize := newFile.Size()
	if oriSize != newSize {
		t.Errorf("expected size to be the same, got: %d, %d", oriSize, newSize)
	}
}
