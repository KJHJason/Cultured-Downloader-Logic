package httpfuncs

import (
	"math/rand/v2"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

// Returns a random time.Duration between the given min and max arguments
//
// Note: If you're not explicitly using the time.Duration type, the arguments will be treated as *nanoseconds*
func GetRandomTime(min, max time.Duration) time.Duration {
	return min + rand.N(max-min)
}

// Converts the given min and max arguments in *milliseconds*
// to time.Duration and returns a random time.Duration between them
func GetRandomTimeIntMs(min, max uint32) time.Duration {
	return GetRandomTime(
		time.Duration(min)*time.Millisecond,
		time.Duration(max)*time.Millisecond,
	)
}

// Returns a random time.Duration between the given min and max arguments in the RetryDelay struct
func GetRandomDelay(delayInfo *RetryDelay) time.Duration {
	return GetRandomTime(delayInfo.Min, delayInfo.Max)
}

// Returns a random time.Duration between the defined min and max delay values in the contants.go file
func GetDefaultRandomDelay() time.Duration {
	return GetRandomTime(constants.MIN_RETRY_DELAY, constants.MAX_RETRY_DELAY)
}
