package httpfuncs

import (
	"math/rand"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

// Returns a random time.Duration between the given min and max arguments
func GetRandomTime(min, max float32) time.Duration {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDelay := min + r.Float32()*(max-min)
	return time.Duration(randomDelay*1000) * time.Millisecond
}

// Returns a random time.Duration between the given min and max arguments in the RetryDelay struct
func GetRandomDelay(delayInfo *RetryDelay) time.Duration {
	return GetRandomTime(delayInfo.Min, delayInfo.Max)
}

// Returns a random time.Duration between the defined min and max delay values in the contants.go file
func GetDefaultRandomDelay() time.Duration {
	return GetRandomTime(constants.MIN_RETRY_DELAY, constants.MAX_RETRY_DELAY)
}
