package tasks

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FileStore is a simple implementation of the TaskInterface which uses a file for persistence.
// FileStore is intended to be a non-performant, non-scalable reference implementation.
type FileStore struct {
	log logr.Logger
	// In memory storage of the tasks
	tasks map[string]*v1alpha1.Task
	// File to back up to.
	fileName string

	mu sync.Mutex
}

// NewFileStore will create a new filestore. Tasks will be persisted to fileName.
// If fileName already exists it should have been created by FileStore.
// The directory for fileName must exist.
func NewFileStore(fileName string, log logr.Logger) (*FileStore, error) {
	if fileName == "" {
		return nil, errors.New("filName is required")
	}

	s := &FileStore{
		log:      log,
		fileName: fileName,
		tasks:    map[string]*v1alpha1.Task{},
	}

	if _, err := os.Stat(fileName); err == nil {
		log.Info("Restoring tasks from file", "file", fileName)
		f, err := os.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to open file: %v", fileName)
		}

		d := json.NewDecoder(f)
		if err := d.Decode(&s.tasks); err != nil {
			return nil, errors.Wrapf(err, "Failed to desirialize map of tasks from file %v", fileName)
		}
	}

	// make sure we can persist to the file.
	s.mu.Lock()
	defer s.mu.Unlock()
	persistErr := s.persistNoLock()

	return s, persistErr
}

func (s *FileStore) Create(ctx context.Context, t *v1alpha1.Task) (*v1alpha1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.GetMetadata().GetName() == "" {
		return t, status.Errorf(codes.InvalidArgument, "name is required")
	}

	if _, ok := s.tasks[t.GetMetadata().GetName()]; ok {
		return t, status.Errorf(codes.AlreadyExists, "Task with name %v; already exists; use Update to make changes", t.GetMetadata().GetName())
	}

	// Generate a unique resourceVersion
	t.Metadata.ResourceVersion = uuid.New().String()
	s.tasks[t.GetMetadata().GetName()] = t
	err := s.persistNoLock()
	return t, err
}

// Get the task with the specified name.
func (s *FileStore) Get(ctx context.Context, name string) (*v1alpha1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	t, ok := s.tasks[name]

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Task %v not found", name)
	}

	return t, nil
}

// Update the specified task. If the task doesn't exist it will be created.
func (s *FileStore) Update(ctx context.Context, t *v1alpha1.Task) (*v1alpha1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := t.GetMetadata().GetName()

	// If task exists check resourceVersion
	if old, ok := s.tasks[name]; ok {
		if old.GetMetadata().GetResourceVersion() != t.GetMetadata().GetResourceVersion() {
			return t, status.Errorf(codes.FailedPrecondition, "ResourceVersion doesn't match; want %v; got %v", old.GetMetadata().GetResourceVersion(), t.GetMetadata().GetResourceVersion())
		}
	}

	// Bump resourceVersion
	t.GetMetadata().ResourceVersion = uuid.New().String()

	s.tasks[name] = t
	err := s.persistNoLock()
	return t, err
}

// Delete the specified task. Delete is a null op if the task doesn't exist
func (s *FileStore) Delete(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[name]; !ok {
		return nil
	}

	delete(s.tasks, name)
	err := s.persistNoLock()
	return err
}

// persistNoLock persists the data to the file. The caller is responsible for acquiring the lock.
func (s *FileStore) persistNoLock() error {
	f, err := os.Create(s.fileName)
	defer DeferIgnoreError(f.Close)

	if err != nil {
		return errors.Wrapf(err, "Failed to create file %v", s.fileName)
	}
	e := json.NewEncoder(f)
	if err := e.Encode(s.tasks); err != nil {
		return errors.Wrapf(err, "Failed to serialize tasks to file: %v", s.fileName)
	}

	return nil
}
