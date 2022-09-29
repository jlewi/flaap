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

// IsDone Returns True iff the task is Done is true and false otherise
func IsDone(t *v1alpha1.Task) bool {
	return GetCondition(t, SUCCEEDED_CONDITION) != v1alpha1.StatusCondition_UNKNOWN
}

// GetCondition returns the value of the specified condition. Returns UNKNOWN if there is no value for the condition.
func GetCondition(t *v1alpha1.Task, name string) v1alpha1.StatusCondition {
	for _, c := range t.GetStatus().GetConditions() {
		// IsDone Returns True iff the task is Done is true and false otherise
		if strings.ToLower(c.GetType()) == name {
			return c.GetStatus()
		}
	}
	return v1alpha1.StatusCondition_UNKNOWN
}
