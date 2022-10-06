package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/zapr"
	"github.com/jlewi/flaap/go/protos/v1alpha1"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultAPIEndpoint = "localhost:8081"
)

func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		Short: "Get a resource",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stdout, "get requires a resource to be specified")
		},
	}
	return cmd
}

func NewGetStatusCmd() *cobra.Command {
	var endpoint string
	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get taskstore status",
		Run: func(cmd *cobra.Command, args []string) {
			log := zapr.NewLogger(zap.L())
			err := func(out io.Writer) error {
				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				conn, err := grpc.Dial(endpoint, opts...)
				if err != nil {
					return errors.Wrapf(err, "Failed to connect to taskstore at %v", endpoint)
				}
				defer conn.Close()

				client := v1alpha1.NewTasksServiceClient(conn)

				popts := protojson.MarshalOptions{
					Multiline: true,
					Indent:    "",
				}

				status, err := client.Status(context.Background(), &v1alpha1.StatusRequest{})
				if err != nil {
					return errors.Wrapf(err, "status request failed")
				}

				b, err := popts.Marshal(status)

				if err != nil {
					return errors.Wrapf(err, "Failed to marshal response to json")
				}
				fmt.Fprintf(out, "%v\n", string(b))
				if err != nil {
					return errors.Wrapf(err, "Failed to get taskstore status")
				}

				return nil
			}(os.Stdout)
			if err != nil {
				log.Error(err, "Error getting taskstore status")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", defaultAPIEndpoint, "The endpoint of the taskstore")
	return cmd
}

func NewGetTasksCmd() *cobra.Command {
	var endpoint string
	var workerId string
	var done bool
	cmd := &cobra.Command{
		Use:   "tasks",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get tasks",
		Run: func(cmd *cobra.Command, args []string) {
			log := zapr.NewLogger(zap.L())
			err := func(out io.Writer) error {
				var opts []grpc.DialOption
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
				conn, err := grpc.Dial(endpoint, opts...)
				if err != nil {
					return errors.Wrapf(err, "Failed to connect to taskstore at %v", endpoint)
				}
				defer conn.Close()

				client := v1alpha1.NewTasksServiceClient(conn)

				name := ""
				if len(args) == 1 {
					name = args[0]
				}

				popts := protojson.MarshalOptions{
					Multiline: true,
					Indent:    "",
				}

				if name == "" {
					// Issue a list request.
					req := &v1alpha1.ListRequest{
						WorkerId: workerId,
						Done:     done,
					}
					log.V(logging.Debug).Info("Issuing list request", "req", req)
					resp, err := client.List(context.Background(), req)

					if err != nil {
						return errors.Wrapf(err, "List request failed")
					}

					b, err := popts.Marshal(resp)

					if err != nil {
						return errors.Wrapf(err, "Failed to marshal response to json")
					}
					fmt.Fprintf(out, "%v\n", string(b))
				} else {
					// Issue a get request.
					req := &v1alpha1.GetRequest{
						Name: name,
					}
					log.V(logging.Debug).Info("Issuing get request", "req", req)
					resp, err := client.Get(context.Background(), req)

					if err != nil {
						return errors.Wrapf(err, "Get request failed")
					}

					b, err := popts.Marshal(resp)

					if err != nil {
						return errors.Wrapf(err, "Failed to marshal response to json")
					}
					fmt.Fprintf(out, "%v\n", string(b))
				}
				return nil
			}(os.Stdout)
			if err != nil {
				log.Error(err, "Error getting the resources")
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", defaultAPIEndpoint, "The endpoint of the taskstore")
	cmd.Flags().StringVarP(&workerId, "workerId", "", "", "Optional; if supplied only list tasks for this worker")
	cmd.Flags().BoolVarP(&done, "done", "", true, "Whether to include done tasks or not")
	return cmd
}