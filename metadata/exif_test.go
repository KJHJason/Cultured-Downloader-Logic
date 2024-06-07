package metadata

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

type testData struct {
	i   int
	ext string
}

func getAllTestImages() []testData {
	return []testData{
		{1, "jpeg"},
		{2, "jpeg"},
		{1, "jpg"},
		{2, "jpg"},
		{1, "png"},
		{2, "png"},
		{1, "webp"},
		{2, "webp"},
	}
}

func getTestI() int {
	return 1
}

func getFileExt() string {
	return "webp"
}

func getTestImgCopyPath(i int, ext string) string {
	return fmt.Sprintf("sample/output/test%d.%s", i, ext)
}

func getTestImgPath(i int, ext string) string {
	return fmt.Sprintf("sample/image%d.%s", i, ext)
}

func getTestMetadata() Metadata {
	return Metadata{
		CreationDate: time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
}

func testWriteExifDataToImage(t *testing.T, i int, ext string) {
	testImgPath := getTestImgPath(i, ext)
	// create a copy of the test image
	src, err := os.Open(testImgPath)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	dstFile := getTestImgCopyPath(i, ext)
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

	err = ChangeFilePathCreationDate(context.Background(), dstFile, getTestMetadata())
	if err != nil {
		t.Fatal(err)
	}
}

// go test -v -timeout 30s -run ^TestWriteExifDataToImage$ github.com/KJHJason/Cultured-Downloader-Logic/metadata
func TestWriteExifDataToImage(t *testing.T) {
	i, ext := getTestI(), getFileExt()
	testWriteExifDataToImage(t, i, ext)
}

// go test -v -run ^TestFileSizeWithoutExifData$ github.com/KJHJason/Cultured-Downloader-Logic/metadata
func TestFileSizeWithoutExifData(t *testing.T) {
	for _, data := range getAllTestImages() {
		testWriteExifDataToImage(t, data.i, data.ext) // generate img with exif data

		sampleImgPath := getTestImgCopyPath(data.i, data.ext) // img with exif data path
		fileSize, err := GetFileSizeWithoutExifData(context.Background(), sampleImgPath)
		if err != nil {
			t.Fatal(err)
		}

		// original img path
		oriFileStat, err := os.Stat(getTestImgPath(data.i, data.ext))
		if err != nil {
			t.Fatal(err)
		}

		oriFileSize := oriFileStat.Size()
		if oriFileSize != fileSize {
			t.Errorf("expected file size to be the same, got: %d, %d for file %q", oriFileSize, fileSize, sampleImgPath)
		}
	}
}

// go test -v -timeout 30s -run ^TestSizeDiff$ github.com/KJHJason/Cultured-Downloader-Logic/metadata
func TestSizeDiff(t *testing.T) {
	TestWriteExifDataToImage(t)

	i := getTestI()
	oriFile, err := os.Stat(getTestImgPath(i, getFileExt()))
	if err != nil {
		t.Fatal(err)
	}
	newFile, err := os.Stat(getTestImgCopyPath(i, getFileExt()))
	if err != nil {
		t.Fatal(err)
	}

	oriSize := oriFile.Size()
	newSize := newFile.Size()
	if oriSize == newSize {
		t.Errorf("expected size to be slightly different, got: %d, %d", oriSize, newSize)
	}

	diff := newSize - oriSize
	t.Logf("size diff: %d", diff)
}
