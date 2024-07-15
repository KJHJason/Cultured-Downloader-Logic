package cf

import (
	"context"
	"testing"
)

// go test -v -run ^TestPullImage$ github.com/KJHJason/Cultured-Downloader-Logic/api/cf
func TestPullImage(t *testing.T) {
	client, err := getDefaultClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if ok, err := hasContainer(ctx, client, IMAGE_NAME); ok {
		t.Log("Image already exists")
		return
	} else if err != nil {
		t.Fatal(err)
	}

	if err := pullImage(ctx, client); err != nil {
		t.Error(err)
	}
}
