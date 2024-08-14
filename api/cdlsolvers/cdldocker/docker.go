package cdldocker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	SUPPORTS_ARM   = false // Currently, Google Chrome does not support ARM
	VERSION        = "0.2.1"
	CONTAINER_NAME = "cdl-solvers"
	IMAGE_NAME     = "kjhjason/" + CONTAINER_NAME + ":" + VERSION
)

func createContainer(ctx context.Context, cli *client.Client, createConfig *utils.ContainerConfigs) (string, error) {
	if createResp, err := utils.CreateContainer(ctx, cli, CONTAINER_NAME, createConfig, SUPPORTS_ARM); err != nil {
		return "", err
	} else {
		return createResp.ID, nil
	}
}

func pullDockerImage(ctx context.Context, cli *client.Client) error {
	if ok, err := utils.HasImage(ctx, cli, IMAGE_NAME); ok {
		return nil
	} else if err != nil {
		return err
	}

	return utils.PullImage(ctx, cli, IMAGE_NAME, SUPPORTS_ARM)
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

func PullDockerImage(ctx context.Context) error {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return err
	}

	if err := pullDockerImage(ctx, cli); err != nil {
		return err
	}
	return nil
}

func dockerCallLogic(ctx context.Context, cli *client.Client, createConfig utils.ContainerConfigs) error {
	var err error
	if err = pullDockerImage(ctx, cli); err != nil {
		return err
	}

	var containerId string
	if containerId, err = createContainer(ctx, cli, &createConfig); err != nil {
		return err
	}

	if err = cli.ContainerStart(ctx, containerId, container.StartOptions{}); err != nil {
		return err
	}
	defer utils.RemoveContainer(ctx, cli, containerId, nil)

	if _, err = utils.WaitForContainer(ctx, cli, containerId); err != nil {
		return err
	}

	return nil
}

func CallDockerImageForCf(ctx context.Context, userAgent string, targetUrl string) ([]*DevToolsCookie, error) {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	dockerLogFileMount, err := getLogPathMount()
	if err != nil {
		return nil, err
	}
	dockerLogFilePath := dockerLogFileMount.Target

	cdlTempDir, err := os.MkdirTemp("", "cdlsolvers-cf-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(cdlTempDir)

	const cookieFilename = "cookie.json"
	cookiePath := filepath.Join(cdlTempDir, cookieFilename)
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

	// using path instead of path/filepath to
	// avoid windows path issues on docker (Linux)
	dockerCookieFilePath := path.Join("/app", utils.GenerateRandomString(12), cookieFilename)
	dockerCookieFileMount := mount.Mount{
		Type:   mount.TypeBind,
		Source: cookieUnixFilePath,
		Target: dockerCookieFilePath,
	}

	cfArgs := newCfArgs(targetUrl, dockerCookieFilePath, userAgent, dockerLogFilePath)
	cmdArgs := cfArgs.parseCmdArgs()

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

	if err := dockerCallLogic(ctx, cli, createConfig); err != nil {
		return nil, err
	}

	cookies, err := parseCookiesFromFile(cookiePath)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}

func CallDockerImageForFantia(ctx context.Context, userAgent string, cookies []*http.Cookie) error {
	cli, err := utils.GetDefaultClient()
	if err != nil {
		return err
	}

	cdlTempDir, err := os.MkdirTemp("", "cdlsolvers-fantia-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(cdlTempDir)

	dockerLogFileMount, err := getLogPathMount()
	if err != nil {
		return err
	}
	dockerLogFilePath := dockerLogFileMount.Target

	cookieParams := convertCookiesToDevToolsCookiesParam(cookies)
	cookiePath, err := makeTempCookieParamFile(cdlTempDir, cookieParams)
	if err != nil {
		return err
	}

	fantiaArgs := newFantiaArgs(userAgent, dockerLogFilePath, cookiePath)
	cmdArgs := fantiaArgs.parseCmdArgs()

	createConfig := utils.ContainerConfigs{
		Config: &container.Config{
			Image: IMAGE_NAME,
			Cmd:   cmdArgs,
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				dockerLogFileMount,
			},
		},
		NetworkingConfig: &network.NetworkingConfig{},
	}

	if err := dockerCallLogic(ctx, cli, createConfig); err != nil {
		return err
	}
	return nil
}
