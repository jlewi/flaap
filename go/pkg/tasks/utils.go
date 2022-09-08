package tasks

import (
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"go.uber.org/zap"
)

const (
	SUCCEEDED_CONDITION = "succeeded"
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

// Returns True iff the task is Done is true and false otherise
func IsDone(t *v1alpha1.Task) bool {
	for _, c := range t.Status.Conditions {
		if strings.ToLower(c.GetType()) == SUCCEEDED_CONDITION {
			if c.GetStatus() == v1alpha1.StatusCondition_UNKNOWN {
				return false
			} else {
				return true
			}
		}
	}
	return false
}
