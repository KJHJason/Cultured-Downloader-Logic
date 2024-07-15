package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Convert the page number to the offset as one page might have x posts.
//
// Usually for paginated results like Pixiv's mobile API (60 per page), checkPixivMax should be set to true.
func ConvertPageNumToOffset(minPageNum, maxPageNum, perPage int) (int, int) {
	// Check for negative page numbers
	if minPageNum < 0 {
		minPageNum = 1
	}
	if maxPageNum < 0 {
		maxPageNum = 1
	}

	// Swap the page numbers if the min is greater than the max
	if minPageNum > maxPageNum {
		minPageNum, maxPageNum = maxPageNum, minPageNum
	}

	minOffset := perPage * (minPageNum - 1)
	maxOffset := perPage * (maxPageNum - minPageNum + 1)
	return minOffset, maxOffset
}

// Returns the min, max, hasMaxNum, and error from the given string of "num" or "min-max"
//
// E.g.
//
//	"1-10" => 1, 10, true, nil
//	"1" => 1, 1, true, nil
//	"" => 1, 1, false, nil (defaults to min = 1, max = inf)
func GetMinMaxFromStr(numStr string) (int, int, bool, error) {
	if numStr == "" {
		// defaults to min = 1, max = inf
		return 1, 1, false, nil
	}

	var err error
	var min, max int
	if strings.Contains(numStr, "-") {
		nums := strings.SplitN(numStr, "-", 2)
		min, err = strconv.Atoi(nums[0])
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert min page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				nums[0],
			)
		}

		max, err = strconv.Atoi(nums[1])
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert max page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				nums[1],
			)
		}

		if min > max {
			min, max = max, min
		}
	} else {
		min, err = strconv.Atoi(numStr)
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				numStr,
			)
		}
		max = min
	}
	return min, max, true, nil
}

// Same as strings.Join([]string, "\n")
func CombineStringsWithNewline(strs ...string) string {
	return strings.Join(strs, "\n")
}

// Checks if the slice of string all matches the given regex pattern
// If strict is true, then all string must match the regex pattern. Otherwise, empty strings are allowed.
//
// Returns true if all matches, false otherwise with the outlier string
func SliceMatchesRegex(regex *regexp.Regexp, slice []string, strict bool) (bool, string) {
	for _, str := range slice {
		if str == "" && !strict {
			continue
		}
		if !regex.MatchString(str) {
			return false, str
		}
	}
	return true, ""
}

// Detects if the given string contains any passwords
func DetectPasswordInText(text string) bool {
	for _, passwordText := range constants.PASSWORD_TEXTS {
		if strings.Contains(text, passwordText) {
			return true
		}
	}

	for _, passwordRegex := range constants.PASSWORD_REGEXES {
		if passwordRegex.MatchString(text) {
			return true
		}
	}
	return false
}

// Detects if the given string contains any GDrive links and logs it if detected
func DetectGDriveLinks(text, postFolderPath string, isUrl, logUrls bool) bool {
	gdriveFilepath := filepath.Join(postFolderPath, constants.GDRIVE_FILENAME)
	containsGDriveLink := false
	if isUrl && strings.HasPrefix(text, constants.GDRIVE_URL) {
		containsGDriveLink = true
	} else if strings.Contains(text, constants.GDRIVE_URL) {
		containsGDriveLink = true
	}

	if !containsGDriveLink {
		return false
	}

	if isUrl {
		gdriveText := fmt.Sprintf(
			"Google Drive link detected: %s\n\n",
			text,
		)
		logger.LogMessageToPath(gdriveText, gdriveFilepath, logger.INFO)
	}
	return true
}

// Detects if the given string contains any other external file hosting providers links and logs it if detected
func DetectOtherExtDLLink(text, postFolderPath string) bool {
	otherExtFilepath := filepath.Join(postFolderPath, constants.OTHER_LINKS_FILENAME)
	for _, extDownloadProvider := range constants.EXTERNAL_DOWNLOAD_PLATFORMS {
		if strings.Contains(text, extDownloadProvider) {
			otherExtText := fmt.Sprintf(
				"Detected a link that points to an external file hosting in post's description:\n%s\n\n",
				text,
			)
			logger.LogMessageToPath(otherExtText, otherExtFilepath, logger.INFO)
			return true
		}
	}
	return false
}

var CachedChromeExecPath string

// Mainly from https://github.com/chromedp/chromedp/blob/ebf842c7bc28db77d0bf4d757f5948d769d0866f/allocate.go#L349-L395
//
// Note: This is not thread-safe
func GetChromeExecPath() (string, error) {
	if CachedChromeExecPath != "" {
		if found, err := exec.LookPath(CachedChromeExecPath); err == nil {
			return found, nil
		}
	}

	if chromeExec := os.Getenv("CHROME_EXECUTABLE"); chromeExec != "" {
		if found, err := exec.LookPath(chromeExec); err == nil {
			CachedChromeExecPath = chromeExec
			return found, nil
		}
	}

	var locations []string
	switch runtime.GOOS {
	case "darwin":
		locations = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	case "windows":
		locations = []string{
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			filepath.Join(os.Getenv("USERPROFILE"), `AppData\Local\Google\Chrome\Application\chrome.exe`),
		}
	default:
		locations = []string{
			"google-chrome",
			"google-chrome-stable",
			"google-chrome-beta",
			"google-chrome-unstable",
			"/usr/bin/google-chrome",
			"/usr/local/bin/chrome",
			"chrome",
		}
	}

	for _, path := range locations {
		if found, err := exec.LookPath(path); err == nil {
			CachedChromeExecPath = path
			return found, nil
		}
	}
	return "", cdlerrors.ErrChromeNotFound
}

func checkPythonExec() bool {
	_, err := exec.LookPath("python")
	return err == nil
}

func checkDockerExec() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func checkDockerDaemonIsRunning() bool {
	cmd := exec.Command("docker", "version")
	PrepareCmdForBgTask(cmd)
	return cmd.Run() == nil
}

func CheckPrerequisites(panicHandler func(msg string)) {
	if _, err := GetChromeExecPath(); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Google Chrome executable not found, please install Google Chrome or set the CHROME_EXECUTABLE environment variable",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}

	useDockerArg := os.Getenv("CDL_CF_USE_DOCKER")
	useDocker := useDockerArg == "1" || useDockerArg == "true"
	if runtime.GOOS == "linux" && !useDocker {
		if !checkPythonExec() {
			panicHandler(
				fmt.Sprintf(
					"error %d: Python executable not found, please install Python and ensure it's in the PATH environment variable",
					cdlerrors.STARTUP_ERROR,
				),
			)
		}
	} else {
		if !checkDockerExec() {
			panicHandler(
				fmt.Sprintf(
					"error %d: Docker executable not found, please install Docker and ensure it's in the PATH environment variable",
					cdlerrors.STARTUP_ERROR,
				),
			)
		}
		if !checkDockerDaemonIsRunning() {
			panicHandler(
				fmt.Sprintf(
					"error %d: Docker is not running, please start Docker daemon",
					cdlerrors.STARTUP_ERROR,
				),
			)
		}
	}
}
