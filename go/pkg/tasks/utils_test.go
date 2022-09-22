package tasks

import (
	"testing"

	"github.com/jlewi/flaap/go/protos/v1alpha1"
)

func Test_is_done(t *testing.T) {
	type testCase struct {
		name       string
		conditions []*v1alpha1.Condition
		expected   bool
	}

	cases := []testCase{
		{
			name: "success",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_TRUE,
				},
			},
			expected: true,
		},
		{
			name: "failure",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_FALSE,
				},
			},
			expected: true,
		},
		{
			name: "unknown",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_UNKNOWN,
				},
			},
			expected: false,
		},
		{
			name: "missing",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "other",
					Status: v1alpha1.StatusCondition_TRUE,
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			task := &v1alpha1.Task{}
			task.Status = &v1alpha1.TaskStatus{}
			task.Status.Conditions = c.conditions

			actual := IsDone(task)

			if actual != c.expected {
				t.Errorf("Got %v; want %v", actual, c.expected)
			}

		})
	}
}

func Test_GetCondition(t *testing.T) {
	type testCase struct {
		descr      string
		conditions []*v1alpha1.Condition
		name       string
		expected   v1alpha1.StatusCondition
	}

	cases := []testCase{
		{
			descr: "success",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_TRUE,
				},
			},
			name:     "succeeded",
			expected: v1alpha1.StatusCondition_TRUE,
		},
		{
			descr: "failure",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_FALSE,
				},
			},
			name:     "succeeded",
			expected: v1alpha1.StatusCondition_FALSE,
		},
		{
			descr: "unknown",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "succeeded",
					Status: v1alpha1.StatusCondition_UNKNOWN,
				},
			},
			name:     "succeeded",
			expected: v1alpha1.StatusCondition_UNKNOWN,
		},
		{
			descr: "missing",
			conditions: []*v1alpha1.Condition{
				{
					Type:   "other",
					Status: v1alpha1.StatusCondition_TRUE,
				},
			},
			name:     "succeeded",
			expected: v1alpha1.StatusCondition_UNKNOWN,
		},
	}

	for _, c := range cases {
		t.Run(c.descr, func(t *testing.T) {
			task := &v1alpha1.Task{}
			task.Status = &v1alpha1.TaskStatus{}
			task.Status.Conditions = c.conditions

			actual := GetCondition(task, c.name)

			if actual != c.expected {
				t.Errorf("Got %v; want %v", actual, c.expected)
			}

		})
	}
}
