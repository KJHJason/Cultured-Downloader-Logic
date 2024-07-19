package utils

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// go test -v -run ^TestPullImage$ github.com/KJHJason/Cultured-Downloader-Logic/api/cf
func TestPullImage(t *testing.T) {
	client, err := GetDefaultClient()
	if err != nil {
		t.Fatal(err)
	}

	const containerName = "cdl-cf"
	const imageName = "kjhjason/" + containerName + ":v0.1.0"
	ctx := context.Background()
	if ok, err := HasImage(ctx, client, imageName); ok {
		t.Log("Image already exists")
		return
	} else if err != nil {
		t.Fatal(err)
	}

	if err := PullImage(ctx, client, imageName); err != nil {
		t.Error(err)
	}

	config := ContainerConfigs{
		Config: &container.Config{
			Image: imageName,
		},
		HostConfig:       &container.HostConfig{},
		NetworkingConfig: &network.NetworkingConfig{},
	}

	var createResp container.CreateResponse
	if createResp, err = CreateContainer(ctx, client, containerName, &config, false); err != nil {
		t.Error(err)
	}
	t.Logf("Container ID: %s", createResp.ID)
}
