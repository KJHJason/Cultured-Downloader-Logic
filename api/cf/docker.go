package cf

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	CONTAINER_NAME = "cdl-cf"
	IMAGE_NAME     = "kjhjason/" + CONTAINER_NAME + ":" + VERSION
)

func createCfContainer(ctx context.Context, cli *client.Client, createConfig *utils.ContainerConfigs) (string, error) {
	if createResp, err := utils.CreateContainer(ctx, cli, CONTAINER_NAME, createConfig); err != nil {
		return "", err
	} else {
		return createResp.ID, nil
	}
}

func pullCfDockerImage(ctx context.Context, cli *client.Client) error {
	if ok, err := utils.HasImage(ctx, cli, IMAGE_NAME); ok {
		return nil
	} else if err != nil {
		return err
	}

	return utils.PullImage(ctx, cli, IMAGE_NAME)
}

func getLogPathMount() (mount.Mount, error) {
	var logMount mount.Mount

	cdlCfLogFilePath := logger.CdlCfLogFilePath
	if !iofuncs.PathExists(cdlCfLogFilePath) {
		// create the log file so that docker doesn't bind it as a directory
		if f, err := os.OpenFile(cdlCfLogFilePath, os.O_RDONLY|os.O_CREATE, constants.DEFAULT_PERMS); err != nil {
			return logMount, fmt.Errorf("failed to create log file: %w", err)
		} else {
			f.Close()
		}
	}
	cdlCfLogUnixFilePath := cdlCfLogFilePath
	if runtime.GOOS == "windows" {
		cdlCfLogUnixFilePath = filepath.ToSlash(cdlCfLogUnixFilePath)
	}

	// using path instead of path/filepath to
	// avoid windows path issues on docker (Linux)
	dockerLogFilePath := path.Join("/app", utils.GenerateRandomString(12), path.Base(cdlCfLogUnixFilePath))
	logMount = mount.Mount{
		Type:   mount.TypeBind,
		Source: cdlCfLogUnixFilePath,
		Target: dockerLogFilePath,
	}
	return logMount, nil
}

func PullCfDockerImage(ctx context.Context) error {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return err
	}

	if err := pullCfDockerImage(ctx, cli); err != nil {
		return err
	}
	return nil
}

func CallDockerImage(ctx context.Context, args CfArgs) (Cookies, error) {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	cdlCfTempDir, err := os.MkdirTemp("", "cdl-cf-")
	if err != nil {
		return nil, err
	}

	const cookieFilename = "cookie.json"
	cookiePath := filepath.Join(cdlCfTempDir, cookieFilename)
	// create the cookie file so that docker doesn't bind it as a directory
	if f, err := os.OpenFile(cookiePath, os.O_RDONLY|os.O_CREATE, constants.DEFAULT_PERMS); err != nil {
		return nil, fmt.Errorf("failed to create cookie file: %w", err)
	} else {
		f.Close()
	}

	cookieUnixFilePath := cookiePath
	if runtime.GOOS == "windows" {
		cookieUnixFilePath = filepath.ToSlash(cookieUnixFilePath)
	}
	defer os.RemoveAll(cdlCfTempDir)

	dockerLogFileMount, err := getLogPathMount()
	if err != nil {
		return nil, err
	}
	dockerLogFilePath := dockerLogFileMount.Target

	// using path instead of path/filepath to
	// avoid windows path issues on docker (Linux)
	dockerCookieFilePath := path.Join("/app", utils.GenerateRandomString(12), cookieFilename)
	dockerCookieFileMount := mount.Mount{
		Type:   mount.TypeBind,
		Source: cookieUnixFilePath,
		Target: dockerCookieFilePath,
	}

	args.BrowserPath = ""
	args.Headless = false
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
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				dockerLogFileMount,
				dockerCookieFileMount,
			},
		},
		NetworkingConfig: &network.NetworkingConfig{},
	}

	if err := pullCfDockerImage(ctx, cli); err != nil {
		return nil, err
	}

	var containerId string
	if containerId, err = createCfContainer(ctx, cli, &createConfig); err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		return nil, err
	}
	defer func() {
		if err := utils.RemoveContainer(ctx, cli, containerId, nil); err != nil {
			logger.LogError(
				fmt.Errorf(
					"error %d: failed to remove container => %w",
					cdlerrors.UNEXPECTED_ERROR, err,
				),
				logger.ERROR,
			)
		}
	}()

	if _, err := utils.WaitForContainer(ctx, cli, containerId); err != nil {
		return nil, err
	}

	cookies, err := parseCookies(cookiePath)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}
