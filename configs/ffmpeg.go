package configs

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

var defaultFfmpegPath string

func init() {
	err := checkFfmpegBinIsValid(context.Background(), "ffmpeg")
	if err == nil {
		defaultFfmpegPath = "ffmpeg"
	}
}

// Returns true if the provided path is a valid FFmpeg binary
func checkFfmpegBinIsValid(ctx context.Context, ffmpegPath string) error {
	if ffmpegPath == "" {
		return errors.New("ffmpeg path is empty")
	}

	_, err := exec.LookPath(ffmpegPath)
	if err != nil {
		return err
	}

	cmdCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// execute the ffmpeg binary to check if it's working
	cmd := exec.CommandContext(cmdCtx, ffmpegPath, "-version")
	utils.PrepareCmdForBgTask(cmd)
	stdout, ffmpegErr := cmd.Output()
	if ffmpegErr != nil {
		return ffmpegErr
	}

	if len(stdout) > 0 && strings.HasPrefix(string(stdout), "ffmpeg version") {
		return nil
	}
	return errors.New("unexpected output from ffmpeg binary, please ensure it is the correct ffmpeg binary")
}

func ValidateFfmpegPathLogic(ctx context.Context, ffmpegPath string) error {
	if defaultFfmpegPath != "" {
		return nil
	}

	_, ffmpegErr := exec.LookPath(ffmpegPath)
	if ffmpegErr != nil {
		return ffmpegErr
	}

	return checkFfmpegBinIsValid(ctx, ffmpegPath)
}

func (c *Config) ValidateFfmpegPathLogic(ctx context.Context) error {
	return ValidateFfmpegPathLogic(ctx, c.FfmpegPath)
}
