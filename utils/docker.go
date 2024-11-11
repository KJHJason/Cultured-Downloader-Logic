package utils

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
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
	logger.MainLogger.Info("Initialising Docker client with Environment variables")
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to initialise Docker client => %v",
			cdlerrors.DOCKER_ERROR,
			err,
		)
		return nil, err
	}
	return client, nil
}

func GetContainerId(ctx context.Context, cli *client.Client, imageName string) (string, error) {
	logger.MainLogger.Infof("Getting container ID for image %s", imageName)
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to list containers => %v",
			cdlerrors.DOCKER_ERROR,
			err,
		)
		return "", err
	}

	for _, cont := range containers {
		if cont.Image == imageName {
			logger.MainLogger.Infof("Container ID for image %s is %s", imageName, cont.ID)
			return cont.ID, nil
		}
	}
	logger.MainLogger.Infof("Container ID for image %s not found", imageName)
	return "", nil
}

func HasImage(ctx context.Context, cli *client.Client, imageName string) (bool, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to list images => %v",
			cdlerrors.DOCKER_ERROR,
			err,
		)
		return false, err
	}
	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}
		for _, tag := range img.RepoTags {
			if tag == imageName {
				logger.MainLogger.Infof("Image %s already exists", imageName)
				return true, nil
			}
		}
	}

	logger.MainLogger.Infof("Image %s does not exist", imageName)
	return false, nil
}

func PullImage(ctx context.Context, cli *client.Client, imageName string, supportsArm bool) error {
	if hasImage, err := HasImage(ctx, cli, imageName); err != nil {
		return err
	} else if hasImage {
		return nil
	}

	pullOptions := image.PullOptions{}
	if supportsArm {
		pullOptions.Platform = GetImagePlatform()
	}

	logger.MainLogger.Infof("Pulling image %s", imageName)
	res, err := cli.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to pull image %s => %v",
			cdlerrors.DOCKER_ERROR,
			imageName,
			err,
		)
		return err
	}
	defer res.Close()

	var output []byte
	if output, err = io.ReadAll(res); err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to read image pull output => %v",
			cdlerrors.DOCKER_ERROR,
			err,
		)
		return err
	}
	logger.MainLogger.Infof("Image %s pulled, response output below...\n%s\n", imageName, string(output))
	return nil
}

type ContainerConfigs struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

func CreateContainer(ctx context.Context, cli *client.Client, containerName string, configs *ContainerConfigs, supportsArm bool) (container.CreateResponse, error) {
	var arch string
	if supportsArm && CheckIsArm() {
		arch = "arm64"
	} else {
		arch = "amd64"
	}

	platformConfig := &ocispec.Platform{
		OS:           "linux",
		Architecture: arch,
	}

	logger.MainLogger.Infof("Creating container %s", containerName)
	resp, err := cli.ContainerCreate(
		ctx,
		configs.Config,
		configs.HostConfig,
		configs.NetworkingConfig,
		platformConfig,
		containerName,
	)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to create container %s => %v",
			cdlerrors.DOCKER_ERROR,
			containerName,
			err,
		)
	}
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

	if err := cli.ContainerRemove(ctx, containerId, options); err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to remove container %s => %v",
			cdlerrors.DOCKER_ERROR,
			containerId,
			err,
		)
		return err
	}
	logger.MainLogger.Infof("Container %s removed", containerId)
	return nil
}

func WaitForContainer(ctx context.Context, cli *client.Client, containerId string) (*container.WaitResponse, error) {
	logger.MainLogger.Infof("Waiting for container %s to stop", containerId)
	statusCh, errCh := cli.ContainerWait(ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		logger.MainLogger.Errorf(
			"error %d: failed to wait for container or container did not exit successfully %s => %v",
			cdlerrors.DOCKER_ERROR,
			containerId,
			err,
		)
		return nil, err
	case waitRes := <-statusCh:
		logger.MainLogger.Infof("Container %s stopped", containerId)
		if waitRes.StatusCode == 0 {
			return &waitRes, nil
		}

		errMsg := fmt.Sprintf(
			"error %d: container %s exited with non-zero status code %d",
			cdlerrors.DOCKER_ERROR,
			containerId,
			waitRes.StatusCode,
		)
		if dockerErrDetails := waitRes.Error; dockerErrDetails != nil && dockerErrDetails.Message != "" {
			errMsg += fmt.Sprintf(
				" with the following error details below...\n%s\n",
				dockerErrDetails.Message,
			)
		}
		logger.MainLogger.Error(errMsg)
		return &waitRes, errors.New(errMsg)
	}
}
