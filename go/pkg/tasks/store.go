package tasks

import (
	"context"

	"github.com/jlewi/flaap/go/protos/v1alpha1"
)

// TaskInterface defines the CRUD interface for the taskstore.
// Modeled on K8s; e.g. https://pkg.go.dev/k8s.io/client-go/kubernetes/typed/core/v1#PodInterface
type TaskInterface interface {
	Create(ctx context.Context, task *v1alpha1.Task) (*v1alpha1.Task, error)
	Get(ctx context.Context, name string) (*v1alpha1.Task, error)
	Update(ctx context.Context, task *v1alpha1.Task) (*v1alpha1.Task, error)
	Delete(ctx context.Context, name string) error
}
