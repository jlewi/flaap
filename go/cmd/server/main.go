// package main provides the entrypoint for the gRPC server
package main

import (
	"fmt"

	"net"
	"os"

	"github.com/go-logr/logr"
	"github.com/jlewi/flaap/go/pkg/tasks"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/go-logr/zapr"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	log logr.Logger
)

func newRootCmd() *cobra.Command {
	var level string
	var jsonLog bool
	var fileName string
	var port int
	rootCmd := &cobra.Command{
		Short: "Run the tasks service",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			newLogger, err := logging.InitLogger(level, !jsonLog)
			if err != nil {
				panic(err)
			}
			log = *newLogger
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := run(port, fileName)
			if err != nil {
				log.Error(err, "Error running grpc service")
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "/tmp/tasks.json", "The file where tasks should be persisted.")
	rootCmd.PersistentFlags().StringVarP(&level, "level", "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", true, "Enable json logging.")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8081, "Port to serve on")
	return rootCmd
}

func run(port int, fileName string) error {
	zc := zap.NewProductionConfig()
	z, _ := zc.Build()
	log := zapr.NewLogger(z)

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

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}
