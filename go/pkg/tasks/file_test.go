package tasks

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/go-logr/zapr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	taskComparer     = cmpopts.IgnoreFields(v1alpha1.Task{}, "state", "sizeCache", "unknownFields")
	metadataComparer = cmpopts.IgnoreFields(v1alpha1.Metadata{}, "state", "sizeCache", "unknownFields")
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

				e := json.NewEncoder(f)

				if err := e.Encode(c.existingTasks); err != nil {
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
					t.Errorf("Expected number of loaded tasks is wrong; Got %v; Want %v", s.tasks, c.existingTasks)
				}

			} else {
				if d := cmp.Diff(c.existingTasks, s.tasks, taskComparer); d != "" {
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
		name string
		task *v1alpha1.Task
		code codes.Code
	}

	cases := []testCase{
		{
			name: "basic",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
			},
			code: codes.OK,
		},
		{
			name: "exists",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task2",
				},
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
			},
			code: codes.InvalidArgument,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Prepopulate the task if this test case is for already exists
			if c.code == codes.AlreadyExists {
				s.tasks[c.task.GetMetadata().GetName()] = c.task
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

			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.tasks[task.GetMetadata().GetName()]; !ok {
				t.Fatalf("Create didn't persist tasks to file.")
			}
		})
	}
}

func Test_Update(t *testing.T) {
	type testCase struct {
		name         string
		input        *v1alpha1.Task
		existingTask *v1alpha1.Task
		code         codes.Code
	}

	cases := []testCase{
		{
			name: "newTask",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
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
			},
			existingTask: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task2",
					ResourceVersion: "1234",
				},
			},
			code: codes.OK,
		},
		{
			name: "existsInvalidResourceVersion",
			input: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task3",
					ResourceVersion: "notamatch",
				},
			},
			existingTask: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name:            "task3",
					ResourceVersion: "1234",
				},
			},
			code: codes.FailedPrecondition,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newFileStore(t)

			// Prepopulate the task if this test case is for already exists
			if c.existingTask != nil {
				s.tasks[c.existingTask.GetMetadata().GetName()] = c.existingTask
			}

			task, updateErr := s.Update(context.Background(), c.input)

			// Handle expected failures
			actualCode := status.Code(updateErr)
			if actualCode != c.code {
				t.Fatalf("Update didn't return expected error code; Got %v; Want %v", actualCode, c.code)
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
			if task.GetMetadata().GetResourceVersion() != s.tasks[task.GetMetadata().GetName()].GetMetadata().GetResourceVersion() {
				t.Fatalf("Update didn't bump ResourceVersion in stored task")
			}

			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.tasks[task.GetMetadata().GetName()]; !ok {
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
				s.tasks[c.task.GetMetadata().GetName()] = c.task
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
		name string
		task *v1alpha1.Task
		code codes.Code
	}

	cases := []testCase{
		{
			name: "ok",
			task: &v1alpha1.Task{
				ApiVersion: "v1",
				Metadata: &v1alpha1.Metadata{
					Name: "task1",
				},
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

			// Prepopulate the task if this test case is for already exists
			taskName := c.task.GetMetadata().GetName()
			if c.task != nil {
				s.tasks[taskName] = c.task
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
			if _, ok := s.tasks[taskName]; ok {
				t.Fatalf("Task was not actually deleted")
			}

			// ensure persistence
			newS, err := NewFileStore(s.fileName, s.log)
			if err != nil {
				t.Fatalf("Failed to create filestore; error: %v", err)
			}

			if _, ok := newS.tasks[taskName]; ok {
				t.Fatalf("Delete didn't persist deletion to file.")
			}
		})
	}
}
