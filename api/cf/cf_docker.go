package cf

import (
	"context"

	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	IMAGE_NAME = "kjhjason/cdl-cf:v0.1.0"
)

func getImagePlatform() string {
	if utils.CheckIsArm() {
		return "linux/arm64"
	}
	return "linux/amd64"
}

func getDefaultClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv)
}

func hasContainer(ctx context.Context, cli *client.Client, containerName string) (bool, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == containerName {
				return true, nil
			}
		}
	}
	return false, nil
}

func pullImage(ctx context.Context, cli *client.Client) error {
	if hasContainer, err := hasContainer(ctx, cli, IMAGE_NAME); err != nil {
		return err
	} else if hasContainer {
		return nil
	}

	pullOptions := image.PullOptions{
		Platform: getImagePlatform(),
	}
	res, err := cli.ImagePull(ctx, IMAGE_NAME, pullOptions)
	if err != nil {
		return err
	}
	defer res.Close()

	// io.Copy(os.Stdout, res)
	return nil
}
