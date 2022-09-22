package tasks

import (
	"context"
	"os"
	"path"
	"testing"

	v0 "github.com/jlewi/flaap/go/protos/tff/v0"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	taskComparer           = cmpopts.IgnoreFields(v1alpha1.Task{}, "state", "sizeCache", "unknownFields")
	taskInputComparer      = cmpopts.IgnoreFields(v1alpha1.TaskInput{}, "state", "sizeCache", "unknownFields")
	anyComparer            = cmpopts.IgnoreFields(anypb.Any{}, "state", "sizeCache", "unknownFields")
	metadataComparer       = cmpopts.IgnoreFields(v1alpha1.Metadata{}, "state", "sizeCache", "unknownFields")
	ignoreResourceVersion  = cmpopts.IgnoreFields(v1alpha1.Metadata{}, "ResourceVersion")
	valueComparer          = cmpopts.IgnoreFields(v0.Value{}, "state", "sizeCache", "unknownFields")
	statusComparer         = cmpopts.IgnoreFields(v1alpha1.TaskStatus{}, "state", "sizeCache", "unknownFields")
	conditionComparer      = cmpopts.IgnoreFields(v1alpha1.Condition{}, "state", "sizeCache", "unknownFields")
	statusResponseComparer = cmpopts.IgnoreFields(v1alpha1.StatusResponse{}, "state", "sizeCache", "unknownFields")
)

func Test_NewFileStore(t *testing.T) {
	type testCase struct {
		name          string
		existingTasks map[string]*v1alpha1.Task
	}

	cases := []testCase{
		{
			name:          "File doesn't exist",
			existingTasks: nil,
		},
		{
			name: "File exists",
			existingTasks: map[string]*v1alpha1.Task{

				"task1": {
					ApiVersion: "v1",
					Kind:       "task",
					Metadata: &v1alpha1.Metadata{
						Name: "task1",
					},
					Input: &v1alpha1.TaskInput{
						Function: []byte("1234abcd"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "testNewFileStore")
			if err != nil {
				t.Fatalf("Failed to create temp directory; error %v", err)
			}
			defer os.RemoveAll(dir)
			fileName := path.Join(dir, "tasks.json")
			if c.existingTasks != nil {
				f, err := os.Create(fileName)
				if err != nil {
					t.Fatalf("Failed to create file: %v, error %v", fileName, err)
				}

				data := &v1alpha1.StoredData{
					Tasks: []*v1alpha1.Task{},
				}
				for _, t := range c.existingTasks {
					data.Tasks = append(data.Tasks, t)
				}

				b, err := protojson.Marshal(data)

				if err != nil {
					t.Fatalf("Failed to marshal StoredData; error: %v", err)
				}
				if _, err := f.Write(b); err != nil {
					t.Fatalf("Failed to serialize existing tasks to file: %v; error %v", fileName, err)
				}
				f.Close()
			}

			s, err := NewFileStore(fileName, zapr.NewLogger(zap.L()))

			if err != nil {
				t.Fatalf("Failed to create FileStore; error %v", err)
			}

			if c.existingTasks == nil {
				if len(c.existingTasks) != 0 {
					t.Errorf("Expected number of loaded tasks is wrong; Got %v; Want %v", s.data.Tasks, c.existingTasks)
				}

			} else {
				if d := cmp.Diff(c.existingTasks, s.data.Tasks, taskComparer, metadataComparer, taskInputComparer, valueComparer, anyComparer); d != "" {
					t.Errorf("Loaded tasks didn't match expected; diff:\n%v", d)
				}
			}

		})
	}
}

// fixture to setup a new filestore or fail the test
func newFileStore(t *testing.T) *FileStore {
	dir, err := os.MkdirTemp("", "testFileStore")
	if err != nil {
		t.Fatalf("Failed to create temp directory; error %v", err)
	}
	fileName := path.Join(dir, "tasks.json")

	s, err := NewFileStore(fileName, zapr.NewLogger(zap.L()))
	if err != nil {
		t.Fatalf("Failed to create FileStore; error %v", err)
	}
	return s
}

func Test_Create(t *testing.T) {
	type testCase struct {
		name               string
		task               *v1alpha1.Task
		code               codes.Code
		expectedUnassigned bool
	}

	cases := []testCase{
		{
			name: "basic",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
				GroupNonce: "1234",
			},
			expectedUnassigned: true,
			code:               codes.OK,
		},
		{
			name: "groupIsAssigned",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
				GroupNonce: "1234",
			},
			expectedUnassigned: false,
			code:               codes.OK,
		},
		{
			name: "exists",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task2",
				},
				GroupNonce: "1234",
			},
			code: codes.AlreadyExists,
		},
		{
			name: "noname",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "",
				},
				GroupNonce: "1234",
			},
			code: codes.InvalidArgument,
		},
		{
			name: "nogroup",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task2",
				},
			},
			code: codes.InvalidArgument,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Prepopulate the task if this test case is for already exists
			if c.code == codes.AlreadyExists {
				s.data.Tasks[c.task.GetMetadata().GetName()] = c.task
			}

			// If this group is already supposed to be assigned set s.data accordingly
			if !c.expectedUnassigned {
				s.data.GroupToWorker[c.task.GetGroupNonce()] = "someworker"
			}

			task, createErr := s.Create(context.Background(), c.task)

			// Handle expected failures
			actualCode := status.Code(createErr)
			if actualCode != c.code {
				t.Fatalf("Create didn't return expected error code; Got %v; Want %v", actualCode, c.code)
			}
			if c.code != codes.OK {
				return
			}

			// Handle expected successes
			if createErr != nil {
				t.Fatalf("Create failed with unexpected error; %+v", createErr)
			}

			if task.GetMetadata().GetResourceVersion() == "" {
				t.Fatalf("Create didn't return non empty ResourceVersion")
			}

			// Ensure task is added to the group list
			groupTasks := s.data.GroupToTaskNames[task.GroupNonce]

			name := c.task.GetMetadata().GetName()
			if groupTasks[len(groupTasks)-1] != name {
				t.Errorf("task %v was not added to s.data.GroupToTaskNames[%v]", name, task.GroupNonce)
			}

			group := c.task.GroupNonce
			// Ensure its added to the unassigned group or not depending on whether the group is already assigned
			// to a worker.

			// This task group is already assigned to a worker so it should not be in the unassigned group
			assignedWorker := s.data.GroupToWorker[group]
			actualUnassigned := assignedWorker == ""
			if actualUnassigned != c.expectedUnassigned {
				t.Errorf("Group %v presence in s.data.UnassignedGroups is wrong; Got %v; want %v", group, actualUnassigned, c.expectedUnassigned)
			}

			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.data.Tasks[task.GetMetadata().GetName()]; !ok {
				t.Fatalf("Create didn't persist tasks to file.")
			}
		})
	}
}

func Test_Update(t *testing.T) {
	type testCase struct {
		name          string
		input         *v1alpha1.Task
		worker        string
		workerToGroup map[string]string
		existingTask  *v1alpha1.Task
		code          codes.Code
	}

	cases := []testCase{
		{
			name: "newTask",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
				GroupNonce: "g1",
			},
			code: codes.OK,
		},
		{
			name: "exists",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Kind:       "updatedField",
				Metadata: &v1alpha1.Metadata{
					Name:            "task2",
					ResourceVersion: "1234",
				},
				GroupNonce: "g1",
			},
			existingTask: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task2",
					ResourceVersion: "1234",
				},
				GroupNonce: "g1",
			},
			worker: "w1",
			workerToGroup: map[string]string{
				"w1": "g1",
			},
			code: codes.OK,
		},
		{
			name: "exists-but-not-assigned-to-that-worker",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Kind:       "updatedField",
				Metadata: &v1alpha1.Metadata{
					Name:            "task2",
					ResourceVersion: "1234",
				},
				GroupNonce: "g1",
			},
			existingTask: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task2",
					ResourceVersion: "1234",
				},
				GroupNonce: "g1",
			},
			worker: "w1",
			code:   codes.FailedPrecondition,
		},
		{
			name: "existsInvalidResourceVersion",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task3",
					ResourceVersion: "notamatch",
				},
				GroupNonce: "g1",
			},
			existingTask: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task3",
					ResourceVersion: "1234",
				},
				GroupNonce: "g1",
			},
			code: codes.FailedPrecondition,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			if c.workerToGroup != nil {
				s.data.WorkerIdToGroupNonce = c.workerToGroup
			}
			// Prepopulate the task if this test case is for already exists
			if c.existingTask != nil {
				s.data.Tasks[c.existingTask.GetMetadata().GetName()] = c.existingTask
			}

			task, updateErr := s.Update(context.Background(), c.input, c.worker)

			// Handle expected failures
			actualCode := status.Code(updateErr)
			if actualCode != c.code {
				t.Fatalf("Update didn't return expected error code; Got %v; Want %v; error %+v", actualCode, c.code, updateErr)
			}

			if c.code != codes.OK {
				return
			}

			// Handle expected successes
			if updateErr != nil {
				t.Fatalf("Update failed with unexpected error; %+v", updateErr)
			}

			// Check resourceVersion got bumped
			if task.GetMetadata().GetResourceVersion() == c.existingTask.GetMetadata().GetResourceVersion() {
				t.Fatalf("Update didn't bump ResourceVersion")
			}

			// Make sure the stored task actually got bumped as well.
			if task.GetMetadata().GetResourceVersion() != s.data.Tasks[task.GetMetadata().GetName()].GetMetadata().GetResourceVersion() {
				t.Fatalf("Update didn't bump ResourceVersion in stored task")
			}

			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.data.Tasks[task.GetMetadata().GetName()]; !ok {
				t.Fatalf("Update didn't persist tasks to file.")
			}
		})
	}
}

func Test_Get(t *testing.T) {
	type testCase struct {
		name string
		task *v1alpha1.Task
		code codes.Code
	}

	cases := []testCase{
		{
			name: "noTask",
			task: nil,
			code: codes.NotFound,
		},
		{
			name: "ok",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task2",
				},
			},
			code: codes.OK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Prepopulate the task if this test case is for already exists
			taskName := c.task.GetMetadata().GetName()
			if c.task != nil {
				s.data.Tasks[c.task.GetMetadata().GetName()] = c.task
			} else {
				// Generate a random taskName to fetch since we want to test task not found
				taskName = "nonExistentTask"
			}

			task, getErr := s.Get(context.Background(), taskName)

			// Ensure expected code
			actualCode := status.Code(getErr)
			if actualCode != c.code {
				t.Fatalf("Get didn't return expected error code; Got %v; Want %v", actualCode, c.code)
			}

			if c.code != codes.OK {
				return
			}

			// Handle expected successes
			if getErr != nil {
				t.Fatalf("Get failed with unexpected error; %+v", getErr)
			}

			if d := cmp.Diff(c.task, task, taskComparer, metadataComparer); d != "" {
				t.Fatalf("Get returned unexpected task; diff:\n%v", d)
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	type testCase struct {
		name            string
		task            *v1alpha1.Task
		code            codes.Code
		existingNameMap map[string][]string
	}

	cases := []testCase{
		{
			name: "ok",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
				GroupNonce: "g1",
			},
			code: codes.OK,
		},
		{
			name: "ok-deletetasksnames",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
				GroupNonce: "g1",
			},
			existingNameMap: map[string][]string{
				"g1": {"task0", "task1", "task2"},
			},
			code: codes.OK,
		},
		{
			name: "nonExistentTask",
			code: codes.OK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			expectedLen := -1
			if c.existingNameMap != nil {
				s.data.GroupToTaskNames = c.existingNameMap
				expectedLen = len(c.existingNameMap[c.task.GroupNonce]) - 1
			}

			// Prepopulate the task if this test case is for already exists
			taskName := c.task.GetMetadata().GetName()
			if c.task != nil {
				s.data.Tasks[taskName] = c.task
			}

			deleteErr := s.Delete(context.Background(), taskName)

			// Handle expected failures
			actualCode := status.Code(deleteErr)
			if actualCode != c.code {
				t.Fatalf("Delete didn't return expected error code; Got %v; Want %v", actualCode, c.code)
			}

			if c.code != codes.OK {
				return
			}

			// Handle expected successes
			if deleteErr != nil {
				t.Fatalf("Delete failed with unexpected error; %+v", deleteErr)
			}

			// Ensure it actually got deleted.
			if _, ok := s.data.Tasks[taskName]; ok {
				t.Fatalf("Task was not actually deleted")
			}

			// Make sure the task got removed from the group to task name map
			if c.existingNameMap != nil {
				actualNames := s.data.GroupToTaskNames[c.task.GroupNonce]
				if len(actualNames) != expectedLen {
					t.Errorf("len(s.data.GroupToTaskNames[%v] Got %v; want %v", c.task.GroupNonce, len(actualNames), expectedLen)
				}

				for _, n := range actualNames {
					if n == c.task.GetMetadata().GetName() {
						t.Errorf("s.data.GroupToTaskNames contains %v but it should have been deleted", c.task.GetMetadata().GetName())
					}
				}
			}
			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.data.Tasks[taskName]; ok {
				t.Fatalf("Delete didn't persist deletion to file.")
			}
		})
	}
}

func newTask(name string, group string, done v1alpha1.StatusCondition) *v1alpha1.Task {
	t := &v1alpha1.Task{
		Metadata: &v1alpha1.Metadata{
			Name: name,
		},
		GroupNonce: group,
		Status: &v1alpha1.TaskStatus{
			Conditions: []*v1alpha1.Condition{
				{
					Type:   SUCCEEDED_CONDITION,
					Status: done,
				},
			},
		},
	}
	return t
}

func Test_List(t *testing.T) {
	type testCase struct {
		name          string
		workerId      string
		includeDone   bool
		GroupToWorker map[string]string
		tasks         []*v1alpha1.Task
		expected      []*v1alpha1.Task
	}

	cases := []testCase{
		{
			name:        "unassigned-group-exclude-done",
			workerId:    "worker1",
			includeDone: false,
			GroupToWorker: map[string]string{
				"group2": "worker2",
			},
			tasks: []*v1alpha1.Task{
				newTask("t1", "group1", v1alpha1.StatusCondition_UNKNOWN),
				newTask("t2", "group1", v1alpha1.StatusCondition_TRUE),
				newTask("t3", "group2", v1alpha1.StatusCondition_TRUE),
			},
			expected: []*v1alpha1.Task{
				newTask("t1", "group1", v1alpha1.StatusCondition_UNKNOWN),
			},
		},
		{
			name:        "unassigned-group-include-done",
			workerId:    "worker1",
			includeDone: true,
			GroupToWorker: map[string]string{
				"group2": "worker2",
			},
			tasks: []*v1alpha1.Task{
				newTask("t1", "group1", v1alpha1.StatusCondition_UNKNOWN),
				newTask("t2", "group1", v1alpha1.StatusCondition_TRUE),
				newTask("t3", "group2", v1alpha1.StatusCondition_TRUE),
			},
			expected: []*v1alpha1.Task{
				newTask("t1", "group1", v1alpha1.StatusCondition_UNKNOWN),
				newTask("t2", "group1", v1alpha1.StatusCondition_TRUE),
			},
		},
		{
			name:        "no-unassigned-group",
			workerId:    "worker1",
			includeDone: true,
			GroupToWorker: map[string]string{
				"group2": "worker2",
			},
			tasks: []*v1alpha1.Task{
				newTask("t3", "group2", v1alpha1.StatusCondition_TRUE),
			},
			expected: []*v1alpha1.Task{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Create existing assignments
			for g, w := range c.GroupToWorker {
				s.data.AssignGroup(g, w)
			}

			// Create tasks
			for _, task := range c.tasks {
				if _, err := s.Create(context.Background(), task); err != nil {
					t.Fatalf("Failed to create task: %v", task.GetMetadata().GetName())
				}
			}

			items, err := s.List(context.Background(), c.workerId, c.includeDone)

			if err != nil {
				t.Fatalf("List failed; error:%v", err)
			}

			if d := cmp.Diff(c.expected, items, taskComparer, metadataComparer, ignoreResourceVersion, statusComparer, conditionComparer); d != "" {
				t.Errorf("Didn't get expected tasks; diff:\n%v", d)
			}
		})
	}
}

func Test_Status(t *testing.T) {
	type testCase struct {
		name          string
		GroupToWorker map[string]string
		tasks         []*v1alpha1.Task
		expected      *v1alpha1.StatusResponse
	}

	cases := []testCase{
		{
			name: "basic",
			GroupToWorker: map[string]string{
				"group2": "worker2",
			},
			tasks: []*v1alpha1.Task{
				newTask("t1", "group1", v1alpha1.StatusCondition_UNKNOWN),
				newTask("t2", "group1", v1alpha1.StatusCondition_TRUE),
				newTask("t3", "group2", v1alpha1.StatusCondition_TRUE),
			},
			expected: &v1alpha1.StatusResponse{
				GroupAssignments: map[string]string{
					"group2": "worker2",
					"group1": "",
				},
				TaskMetrics: map[string]int32{
					v1alpha1.StatusCondition_UNKNOWN.String(): 1,
					v1alpha1.StatusCondition_TRUE.String():    2,
					v1alpha1.StatusCondition_FALSE.String():   0,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Create existing assignments
			for g, w := range c.GroupToWorker {
				s.data.AssignGroup(g, w)
			}

			// Create tasks
			for _, task := range c.tasks {
				if _, err := s.Create(context.Background(), task); err != nil {
					t.Fatalf("Failed to create task: %v", task.GetMetadata().GetName())
				}
			}

			actual, err := s.Status(context.Background(), &v1alpha1.StatusRequest{})

			if err != nil {
				t.Fatalf("Status failed; error:%v", err)
			}

			if d := cmp.Diff(c.expected, actual, statusResponseComparer); d != "" {
				t.Errorf("Didn't get expected tasks; diff:\n%v", d)
			}
		})
	}
}
