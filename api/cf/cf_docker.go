package cf

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	CONTAINER_NAME = "cdl-cf"
	IMAGE_NAME     = "kjhjason/" + CONTAINER_NAME + ":v0.1.0"
)

func createCfContainer(ctx context.Context, cli *client.Client, createConfig *utils.ContainerConfigs) (string, error) {
	if createResp, err := utils.CreateContainer(ctx, cli, CONTAINER_NAME, createConfig); err != nil {
		return "", err
	} else {
		return createResp.ID, nil
	}
}

func pullAndCreateCfDockerImage(ctx context.Context, cli *client.Client, createConfig *utils.ContainerConfigs) (string, error) {
	if ok, err := utils.HasImage(ctx, cli, IMAGE_NAME); ok {
		return createCfContainer(ctx, cli, createConfig)
	} else if err != nil {
		return "", err
	}

	if err := utils.PullImage(ctx, cli, IMAGE_NAME); err != nil {
		return "", err
	}
	return createCfContainer(ctx, cli, createConfig)
}

func CallCfDockerImage(ctx context.Context, args CfArgs) (Cookies, error) {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	cdlCfTempDir, err := os.MkdirTemp("", "cdl-cf-")
	if err != nil {
		return nil, err
	}
	cdlCfTempUnixPath := filepath.ToSlash(cdlCfTempDir)
	cdlCfLogUnixFilePath := filepath.ToSlash(logger.CdlCfLogFilePath)

	const cookieFilename = "cookie.json"
	cookiePath := filepath.Join(cdlCfTempDir, cookieFilename)
	defer os.RemoveAll(cdlCfTempDir)

	// using path instead of path/filepath to
	// avoid windows path issues on docker (Linux)
	dockerDir := path.Join("/app", utils.GenerateRandomString(12))

	dockerLogFilePath := path.Join(dockerDir, "logs", path.Base(cdlCfLogUnixFilePath))
	dockerLogDirPath := path.Dir(dockerLogFilePath)

	dockerCookieFilePath := path.Join(dockerDir, "cookies", cookieFilename)
	dockerCookieDirPath := path.Dir(dockerCookieFilePath)

	cmdArgs := args.ParseCmdArgs()
	cmdArgs = append(cmdArgs,
		"--os-name", runtime.GOOS,
		"--cookie-path", dockerCookieFilePath,
		"--log-path", dockerLogFilePath,
		"--virtual-display",
		// yes, it is hardcoded mainly to make the docker
		// image harder to run for people without dev knowledge
		"--app-key", "fzN9Hvkb9s+mwPGCDd5YFnLiqKx8WhZfWoZE5nZC",
	)

	createConfig := utils.ContainerConfigs{
		Config: &container.Config{
			Image: IMAGE_NAME,
			Cmd:   cmdArgs,
			Volumes: map[string]struct{}{
				cdlCfTempUnixPath + ":" + dockerCookieDirPath:           {},
				path.Dir(cdlCfLogUnixFilePath) + ":" + dockerLogDirPath: {},
			},
		},
		HostConfig:       &container.HostConfig{},
		NetworkingConfig: &network.NetworkingConfig{},
	}

	var containerId string
	if containerId, err = pullAndCreateCfDockerImage(ctx, cli, &createConfig); err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		return nil, err
	}

	cookies, err := parseCookies(cookiePath)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}
