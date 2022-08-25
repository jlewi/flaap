package tasks

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
)

// Server implments the gRPC interface for tasks.
type Server struct {
	tasks TaskInterface
	log   logr.Logger
	v1alpha1.UnimplementedTasksServiceServer
}

// NewServer constructs a new server.
func NewServer(tasks TaskInterface, log logr.Logger) (*Server, error) {
	return &Server{
		tasks: tasks,
		log:   log,
	}, nil
}

func (s *Server) Create(ctx context.Context, req *v1alpha1.CreateRequest) (*v1alpha1.CreateResponse, error) {
	t, err := s.tasks.Create(ctx, req.GetTask())
	return &v1alpha1.CreateResponse{Task: t}, err
}

func (s *Server) Get(ctx context.Context, req *v1alpha1.GetRequest) (*v1alpha1.GetResponse, error) {
	t, err := s.tasks.Get(ctx, req.GetName())
	return &v1alpha1.GetResponse{Task: t}, err
}

func (s *Server) Update(ctx context.Context, req *v1alpha1.UpdateRequest) (*v1alpha1.UpdateResponse, error) {
	t, err := s.tasks.Create(ctx, req.GetTask())
	return &v1alpha1.UpdateResponse{Task: t}, err
}

func (s *Server) Delete(ctx context.Context, req *v1alpha1.DeleteRequest) (*v1alpha1.DeleteResponse, error) {
	err := s.tasks.Delete(ctx, req.GetName())
	return &v1alpha1.DeleteResponse{}, err
}
