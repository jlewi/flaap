package tasks

import (
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// IgnoreError is a helper function to deal with errors.
func IgnoreError(err error) {
	if err == nil {
		return
	}
	log := zapr.NewLogger(zap.L())
	log.Error(err, "Unexpected error occurred")
}

// DeferIgnoreError is a helper function to ignore errors returned by functions called with defer.
func DeferIgnoreError(f func() error) {
	IgnoreError(f())
}
