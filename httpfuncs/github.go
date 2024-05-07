package httpfuncs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

var (
	ErrProcessLatestVer = fmt.Errorf(
		"github error %d: unable to process the latest version",
		errs.DEV_ERROR,
	)
	ErrProcessVer = fmt.Errorf(
		"github error %d: unable to process the current version",
		errs.DEV_ERROR,
	)
)

func processVer(apiResVer string) (*versionInfo, error) {
	// split the version string by "."
	ver := strings.Split(apiResVer, ".")
	if len(ver) != 3 {
		return nil, ErrProcessLatestVer
	}

	// convert the version string to int
	verSlice := make([]int, 3)
	for i, v := range ver {
		verInt, err := strconv.Atoi(v)
		if err != nil {
			return nil, ErrProcessLatestVer
		}
		verSlice[i] = verInt
	}

	return &versionInfo{
		Major: verSlice[0],
		Minor: verSlice[1],
		Patch: verSlice[2],
	}, nil
}

// check if the latest version is greater than the current version.
// returns true if the current version is outdated.
func checkIfVerIsOutdated(curVer *versionInfo, latestVer *versionInfo) bool {
	if latestVer.Major > curVer.Major {
		return true
	}

	if latestVer.Major == curVer.Major {
		if latestVer.Minor > curVer.Minor {
			return true
		}

		if latestVer.Minor == curVer.Minor {
			if latestVer.Patch > curVer.Patch {
				return true
			}
		}
	}
	return false
}

// check for the latest version of the program
func CheckVer(repo string, ver string, showProg bool, progBar progress.ProgressBar) (bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	if !constants.GITHUB_VER_REGEX.MatchString(ver) {
		return false, fmt.Errorf(
			"github error %d: unable to process the current version, %q",
			errs.DEV_ERROR,
			ver,
		)
	}

	hasProgBar := showProg && progBar != nil
	if hasProgBar {
		progBar.UpdateBaseMsg("Checking for the latest version...")
		progBar.UpdateErrorMsg("Failed to check for the latest version, please refer to the logs for more details...")
		progBar.Start()
	}

	res, err := CallRequest(
		&RequestArgs{
			Url:         url,
			Method:      "GET",
			Timeout:     5,
			CheckStatus: false,
			Http3:       false,
			Http2:       true,
		},
	)
	if err != nil || res.StatusCode != 200 {
		errMsg := fmt.Errorf(
			"github error %d: unable to check for the latest version",
			errs.CONNECTION_ERROR,
		)
		if err != nil {
			errMsg = fmt.Errorf("%w, more info => %w", errMsg, err)
		}

		if showProg && progBar != nil {
			progBar.Stop(true)
		}
		return false, errMsg
	}

	var apiRes GithubApiRes
	if err := LoadJsonFromResponse(res, &apiRes); err != nil {
		errMsg := fmt.Sprintf(
			"github error %d: unable to marshal the response from the API into an interface",
			errs.UNEXPECTED_ERROR,
		)
		if hasProgBar {
			progBar.Stop(true)
		}
		return false, errors.New(errMsg)
	}

	latestVer, err := processVer(apiRes.TagName)
	if err != nil {
		errMsg := fmt.Sprintf(
			"github error %d: unable to process the latest version",
			errs.UNEXPECTED_ERROR,
		)
		if hasProgBar {
			progBar.UpdateErrorMsg(errMsg)
			progBar.Stop(true)
		}
		return false, err
	}

	programVer, err := processVer(ver)
	if err != nil {
		errMsg := fmt.Sprintf(
			"error %d: unable to process the program version",
			errs.DEV_ERROR,
		)
		if hasProgBar {
			progBar.UpdateErrorMsg(errMsg)
			progBar.Stop(true)
		}
		return false, ErrProcessVer
	}

	outdated := checkIfVerIsOutdated(programVer, latestVer)
	if hasProgBar {
		if outdated {
			progBar.UpdateErrorMsg(
				fmt.Sprintf(
					"Warning: this program is outdated, the latest version %q is available at %s",
					apiRes.TagName,
					apiRes.HtmlUrl,
				),
			)
		} else {
			progBar.UpdateSuccessMsg("This program is up to date!")
		}
		progBar.Stop(outdated)
	}
	return outdated, nil
}
