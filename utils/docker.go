package utils

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func GetImagePlatform() string {
	if CheckIsArm() {
		return "linux/arm64"
	}
	return "linux/amd64"
}

func GetDefaultClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv)
}

func GetContainerId(ctx context.Context, cli *client.Client, imageName string) (string, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, cont := range containers {
		if cont.Image == imageName {
			return cont.ID, nil
		}
	}
	return "", nil
}

func HasImage(ctx context.Context, cli *client.Client, imageName string) (bool, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}
		for _, tag := range img.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}
	return false, nil
}

func PullImage(ctx context.Context, cli *client.Client, imageName string) error {
	if hasImage, err := HasImage(ctx, cli, imageName); err != nil {
		return err
	} else if hasImage {
		return nil
	}

	pullOptions := image.PullOptions{
		Platform: GetImagePlatform(),
	}
	res, err := cli.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return err
	}
	defer res.Close()
	return nil
}

type ContainerConfigs struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

func CreateContainer(ctx context.Context, cli *client.Client, containerName string, configs *ContainerConfigs) (container.CreateResponse, error) {
	var arch string
	if CheckIsArm() {
		arch = "arm64"
	} else {
		arch = "amd64"
	}

	platformConfig := &ocispec.Platform{
		OS:           "linux",
		Architecture: arch,
	}

	resp, err := cli.ContainerCreate(
		ctx,
		configs.Config,
		configs.HostConfig,
		configs.NetworkingConfig,
		platformConfig,
		containerName,
	)
	return resp, err
}

func RemoveContainer(ctx context.Context, cli *client.Client, containerId string, rmOptions *container.RemoveOptions) error {
	var options container.RemoveOptions
	if rmOptions == nil {
		options = container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}
	} else {
		options = *rmOptions
	}
	return cli.ContainerRemove(ctx, containerId, options)
}

func WaitForContainer(ctx context.Context, cli *client.Client, containerId string) (*container.WaitResponse, error) {
	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return nil, err
	case waitRes := <-statusCh:
		return &waitRes, nil
	}
}
