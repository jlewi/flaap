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
	// Update departs from K8s convention because we need to provide the ID of the worker issuing the update
	// request. In principle that should be sent via a sidechannel (e.g. as part of Authn)  so it wouldn't
	// be in the body of the request.
	Update(ctx context.Context, task *v1alpha1.Task, worker string) (*v1alpha1.Task, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context, workerId string, includeDone bool) ([]*v1alpha1.Task, error)
	Status(ctx context.Context, req *v1alpha1.StatusRequest) (*v1alpha1.StatusResponse, error)
}
