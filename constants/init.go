package constants

import (
	"fmt"
	"runtime"
)

func init() {
	var userAgent = map[string]string{
		"linux":   "Mozilla/5.0 (X11; Linux x86_64)",
		"darwin":  "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6)",
		"windows": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	}
	userAgentOS, ok := userAgent[runtime.GOOS]
	if !ok {
		panic(
			fmt.Errorf(
				"error %d: Failed to get user agent OS as your OS, %q, is not supported",
				OS_ERROR,
				runtime.GOOS,
			),
		)
	}
	USER_AGENT = fmt.Sprintf("%s AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36", userAgentOS)
}
