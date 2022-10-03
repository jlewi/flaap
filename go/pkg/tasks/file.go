package tasks

import (
	"context"
	"os"
	"sync"

	"github.com/jlewi/p22h/backend/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"

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
	log  logr.Logger
	data *StoredData
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
		data:     newStoredData(log),
	}

	if _, err := os.Stat(fileName); err == nil {
		log.Info("Restoring tasks from file", "file", fileName)
		b, err := os.ReadFile(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to read file: %v", fileName)
		}

		data := &v1alpha1.StoredData{}

		if err := protojson.Unmarshal(b, data); err != nil {
			return nil, errors.Wrapf(err, "Failed to desirialize map of tasks from file %v", fileName)
		}

		for _, t := range data.GetTasks() {
			s.data.Tasks[t.GetMetadata().GetName()] = t
			group := t.GetGroupNonce()
			if _, ok := s.data.GroupToTaskNames[group]; !ok {
				s.data.GroupToTaskNames[group] = []string{}
			}
			s.data.GroupToTaskNames[group] = append(s.data.GroupToTaskNames[group], t.GetMetadata().GetName())
		}

		if data.GetGroupAssignments() != nil {
			s.data.GroupToWorker = data.GetGroupAssignments()

			for g, w := range s.data.GroupToWorker {
				s.data.WorkerIdToGroupNonce[w] = g
			}
		} else {
			s.data.GroupToWorker = map[string]string{}
			s.data.WorkerIdToGroupNonce = map[string]string{}
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

	name := t.GetMetadata().GetName()
	group := t.GetGroupNonce()

	if group == "" {
		return t, status.Errorf(codes.InvalidArgument, "group nonce is required")
	}

	if _, ok := s.data.Tasks[name]; ok {
		return t, status.Errorf(codes.AlreadyExists, "Task with name %v; already exists; use Update to make changes", name)
	}

	// Generate a unique resourceVersion
	t.Metadata.ResourceVersion = uuid.New().String()
	s.log.Info("Creating task", "task", name, "numTasks", len(s.data.Tasks))
	s.data.Tasks[name] = t

	if g, ok := s.data.GroupToTaskNames[group]; ok {
		// We have seen this group before
		s.log.V(logging.Debug).Info("Adding task to group", "task", name, "group", group)
		s.data.GroupToTaskNames[group] = append(g, name)
	} else {
		s.data.GroupToTaskNames[group] = []string{name}
	}

	if _, ok := s.data.GroupToWorker[group]; !ok {
		s.log.V(logging.Debug).Info("Creating new, unassigned task group", "task", name, "group", group)
		s.data.GroupToWorker[group] = ""
	}

	err := s.persistNoLock()
	// TODO(jeremy): log this at debug level
	s.log.Info("Done creating task", "task", name, "numTasks", len(s.data.Tasks))
	return t, err
}

// Get the task with the specified name.
func (s *FileStore) Get(ctx context.Context, name string) (*v1alpha1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	t, ok := s.data.Tasks[name]

	if !ok {
		return nil, status.Errorf(codes.NotFound, "Task %v not found", name)
	}

	return t, nil
}

// Update the specified task. If the task doesn't exist it will be created
func (s *FileStore) Update(ctx context.Context, t *v1alpha1.Task, worker string) (*v1alpha1.Task, error) {
	log := s.log.WithValues("name", t.GetMetadata().GetName(), "worker", worker)
	log.Info("update task")
	name := t.GetMetadata().GetName()
	exists := func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		_, exists := s.data.Tasks[name]
		return exists
	}()

	// If task doesn't exist create it
	if !exists {
		if worker != "" {
			return t, status.Errorf(codes.FailedPrecondition, "Update non-existing task %v failed. When update is used to create new tasks worker should not be set because workers are not expected to create tasks", t.GetMetadata().GetName())
		}
		return s.Create(ctx, t)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	old := s.data.Tasks[name]
	if old.GetMetadata().GetResourceVersion() != t.GetMetadata().GetResourceVersion() {
		s.log.Info("Task update failed; resourceversion doesn't match", "task", name, "want", old.GetMetadata().GetResourceVersion(), "got", t.GetMetadata().GetResourceVersion())
		return t, status.Errorf(codes.FailedPrecondition, "ResourceVersion doesn't match; want %v; got %v", old.GetMetadata().GetResourceVersion(), t.GetMetadata().GetResourceVersion())
	}

	// Make sure this worker has been assigned to this task otherwise fail this update.
	// NB. We check worker id against the old group because we don't want the clients to be able to change the group id
	// and claim tasks they aren't assigned
	expected, ok := s.data.WorkerIdToGroupNonce[worker]
	if !ok || expected != old.GetGroupNonce() {
		s.log.Info("Task update failed; task not assigned to worker", "worker", worker, "task", name)
		return t, status.Errorf(codes.FailedPrecondition, "Worker %v can't update task %v; this task has not been assigned to that worker", worker, name)
	}

	// Check immutable fields
	if t.GetGroupNonce() != old.GetGroupNonce() {
		return t, status.Errorf(codes.FailedPrecondition, "Group nonce is immutable")
	}

	// TODO(jeremy) We should verify TaskInput hasn't changed

	// Bump resourceVersion
	t.GetMetadata().ResourceVersion = uuid.New().String()

	s.data.Tasks[name] = t

	err := s.persistNoLock()
	return t, err
}

// Delete the specified task. Delete is a null op if the task doesn't exist
func (s *FileStore) Delete(ctx context.Context, name string) error {
	log := s.log.WithValues("name", name)
	log.Info("delete task")
	s.mu.Lock()
	defer s.mu.Unlock()

	old, ok := s.data.Tasks[name]

	if !ok {
		return nil
	}

	delete(s.data.Tasks, name)

	if groupTasks, ok := s.data.GroupToTaskNames[old.GetGroupNonce()]; ok {
		cap := len(groupTasks) - 1
		if cap < 0 {
			cap = 0
		}
		newTasks := make([]string, 0, cap)

		for _, n := range groupTasks {
			if n != name {
				newTasks = append(newTasks, n)
			}
		}

		s.data.GroupToTaskNames[old.GetGroupNonce()] = newTasks
	}
	err := s.persistNoLock()
	return err
}

// List returns tasks for this worker. If the worker isn't already assigned a task group
// one is assigned.
func (s *FileStore) List(ctx context.Context, workerId string, includeDone bool) ([]*v1alpha1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.data.WorkerIdToGroupNonce[workerId]
	log := s.log.WithValues("workerId", workerId)
	if !ok || group == "" {
		// Check for unassigned group
		for k, v := range s.data.GroupToWorker {
			if v == "" {
				log.Info("Assigning worker to group", "group", k)
				group = k
				s.data.GroupToWorker[group] = workerId
				s.data.WorkerIdToGroupNonce[workerId] = group
				break
			}
		}
	}

	if group == "" {
		s.log.Info("Can't assign group to worker; no unassigned groups", "workerId", workerId)
		return []*v1alpha1.Task{}, nil
	}

	names, ok := s.data.GroupToTaskNames[group]

	if !ok {
		return []*v1alpha1.Task{}, status.Errorf(codes.Internal, "Group %v had no tasks listed", group)
	}

	results := []*v1alpha1.Task{}

	for _, n := range names {
		t, ok := s.data.Tasks[n]
		if !ok {
			return []*v1alpha1.Task{}, status.Errorf(codes.Internal, "Task %v is listed in group %v but not found in task store", n, group)
		}

		// Only include the task if either we include done takss or the task isn't done
		if includeDone || !IsDone(t) {
			results = append(results, t)
		}
	}

	return results, nil
}

// Status returns status information about the taskstore.
// N.B this method blocks other updates and is potentially expensive.
func (s *FileStore) Status(ctx context.Context, req *v1alpha1.StatusRequest) (*v1alpha1.StatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp := &v1alpha1.StatusResponse{
		GroupAssignments: make(map[string]string),
	}

	// TODO(jeremy): Is there a more efficient way to copy this data?
	for g, w := range s.data.GroupToWorker {
		resp.GroupAssignments[g] = w
	}

	resp.TaskMetrics = map[string]int32{}

	for n := range v1alpha1.StatusCondition_value {
		resp.TaskMetrics[n] = 0
	}

	for _, t := range s.data.Tasks {
		done := GetCondition(t, SUCCEEDED_CONDITION)
		resp.TaskMetrics[done.String()] = resp.TaskMetrics[done.String()] + 1
	}
	return resp, nil
}

// persistNoLock persists the data to the file. The caller is responsible for acquiring the lock.
func (s *FileStore) persistNoLock() error {
	f, err := os.Create(s.fileName)
	defer DeferIgnoreError(f.Close)
	if err != nil {
		return errors.Wrapf(err, "Failed to create file %v", s.fileName)
	}

	data := &v1alpha1.StoredData{}
	s.log.V(logging.Debug).Info("Initializing storage array", "size", len(s.data.Tasks))
	data.Tasks = make([]*v1alpha1.Task, 0, len(s.data.Tasks))

	for _, t := range s.data.Tasks {
		data.Tasks = append(data.Tasks, t)
	}

	data.GroupAssignments = s.data.GroupToWorker

	b, err := protojson.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal data to serialize")
	}

	if _, err := f.Write(b); err != nil {
		return errors.Wrapf(err, "Failed to serialize tasks to file: %v", s.fileName)
	}

	return nil
}

// StoredData represents all the data to be stored and serialized in the file.
// TODO(jeremy): Should we make this a proto?
type StoredData struct {
	log logr.Logger
	// the tasks
	Tasks map[string]*v1alpha1.Task

	// Group to worker is a mapping from groups to workers; empty means its unassigned.
	// It is the inverse of WorkerIdToGroupNonce
	GroupToWorker map[string]string

	// Mapping from a worker id to the group nonce
	WorkerIdToGroupNonce map[string]string

	// Should we make this a list of only incomplete tasks?
	GroupToTaskNames map[string][]string
}

// AssignGroup Assigns the group to the worker. Overrides existing assignments
func (s *StoredData) AssignGroup(group string, workerId string) {
	s.GroupToWorker[group] = workerId
	s.WorkerIdToGroupNonce[workerId] = group
}

func newStoredData(log logr.Logger) *StoredData {
	return &StoredData{
		log:                  log,
		Tasks:                map[string]*v1alpha1.Task{},
		GroupToWorker:        map[string]string{},
		WorkerIdToGroupNonce: map[string]string{},
		// GroupToTaskNames is a mapping from a group to tasks in that group.
		// Makes it easy to lookup tasks for a given worker.
		GroupToTaskNames: map[string][]string{},
	}
}
