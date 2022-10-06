package commands

import (
	"fmt"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"net"

	"github.com/jlewi/flaap/go/pkg/tasks"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"github.com/pkg/errors"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"
)

// RunServer runs the server on the given port
func RunServer(port int, fileName string) error {
	log := zapr.NewLogger(zap.L())

	fileStore, err := tasks.NewFileStore(fileName, log)

	if err != nil {
		return err
	}

	tasksServer, err := tasks.NewServer(fileStore, log)

	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	v1alpha1.RegisterTasksServiceServer(grpcServer, tasksServer)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	// Register health check.
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return errors.Wrapf(err, "failed to listen: %v", err)
	}
	log.Info("Starting grpc service", "port", port)
	return grpcServer.Serve(lis)
}
