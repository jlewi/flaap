package commands

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"github.com/jlewi/flaap/go/pkg/auth"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
	"io"
	"os"
	"time"

	"google.golang.org/grpc/credentials"

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

	cmd.AddCommand(NewGetTasksCmd())
	cmd.AddCommand(NewGetStatusCmd())
	return cmd
}

func NewGetStatusCmd() *cobra.Command {
	grpcFlags := &GRPCClientFlags{}
	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get taskstore status",
		Run: func(cmd *cobra.Command, args []string) {
			log := zapr.NewLogger(zap.L())
			err := func(out io.Writer) error {
				conn, err := grpcFlags.NewConn()
				if err != nil {
					return err
				}
				defer conn.Close()

				client := v1alpha1.NewTasksServiceClient(conn)

				popts := protojson.MarshalOptions{
					Multiline: true,
					Indent:    "",
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()

				// https://cloud.google.com/trace/docs/setup#force-trace
				// 32 hexadecimal characters = 16 bytes
				traceId, err := randomHex(16)
				if err != nil {
					return errors.Wrapf(err, "Failed to generate traceId")
				}
				traceVal := fmt.Sprintf("%v/1;o=1", traceId)
				md := metadata.Pairs("X-Cloud-Trace-Context", traceVal)
				ctx = metadata.NewOutgoingContext(ctx, md)
				status, err := client.Status(ctx, &v1alpha1.StatusRequest{})
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

	grpcFlags.AddFlags(cmd)
	return cmd
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewGetTasksCmd() *cobra.Command {
	var workerId string
	var done bool

	grpcFlags := &GRPCClientFlags{}

	cmd := &cobra.Command{
		Use:   "tasks",
		Args:  cobra.MaximumNArgs(1),
		Short: "Get tasks",
		Run: func(cmd *cobra.Command, args []string) {
			log := zapr.NewLogger(zap.L())
			err := func(out io.Writer) error {
				conn, err := grpcFlags.NewConn()
				if err != nil {
					return err
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

	cmd.Flags().StringVarP(&workerId, "workerId", "", "", "Optional; if supplied only list tasks for this worker")
	cmd.Flags().BoolVarP(&done, "done", "", true, "Whether to include done tasks or not")

	grpcFlags.AddFlags(cmd)
	return cmd
}

type GRPCClientFlags struct {
	UseTLS     bool
	RootCA     string
	SkipVerify bool
	ServerName string
	Endpoint   string
	Token      string
}

func (f *GRPCClientFlags) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Endpoint, "endpoint", "e", defaultAPIEndpoint, "The endpoint of the taskstore")
	cmd.Flags().BoolVarP(&f.UseTLS, "use-tls", "", true, "Whether to use TLS to connect to the server")
	cmd.Flags().StringVarP(&f.RootCA, "root-ca", "", "", "CA file to use for validating client certs")
	cmd.Flags().StringVarP(&f.Token, "auth-token", "", "", "Authorization token to use.")
	cmd.Flags().StringVarP(&f.ServerName, "server-name", "", "", "The servername to use to validate the certificate")
	cmd.Flags().BoolVarP(&f.SkipVerify, "insecure-skip-verify", "", false, "Whether to verify the server's certificate")
}

// NewConn creates a new connection with the given flogs
func (f *GRPCClientFlags) NewConn() (*grpc.ClientConn, error) {
	log := zapr.NewLogger(zap.L())
	var opts []grpc.DialOption

	var creds credentials.TransportCredentials

	if f.UseTLS {
		capool := x509.NewCertPool()

		if f.RootCA != "" {
			log.Info("Reading root CA", "file", f.RootCA)

			ca, err := os.ReadFile(f.RootCA)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to read rootCA file: %v", f.RootCA)
			}
			if !capool.AppendCertsFromPEM(ca) {
				return nil, errors.Errorf("can't add CA certs to pool")
			}
		}
		creds = credentials.NewTLS(&tls.Config{
			// Certificates: []tls.Certificate{cert},
			RootCAs:            capool,
			InsecureSkipVerify: f.SkipVerify,
			ServerName:         f.ServerName,
			MinVersion:         tls.VersionTLS13,
		})
	} else {
		creds = insecure.NewCredentials()
	}

	// Create channel credentials. This basically sets up TLS for the channel.
	opts = append(opts, grpc.WithTransportCredentials(creds))

	// Create call channels; this will attach a JWT to each request.
	// This requires TLS to be enabled because we don't want to send bearer tokens over http.
	ts, err := f.NewOIDCTokenSource()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create OIDC token source")
	}
	opts = append(opts, grpc.WithPerRPCCredentials(ts))

	conn, err := grpc.Dial(f.Endpoint, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to connect to taskstore at %v", f.Endpoint)
	}
	return conn, nil
}

// NewOIDCTokenSource returns a gRPC TokenSource that uses a JWT as the token.
// A gRPC TokenSource wraps an oauth2 token source but implements grpc's credentials.PerRPCCredentials interface
// which allows to inject credentials on each call.
func (f *GRPCClientFlags) NewOIDCTokenSource() (*oauth.TokenSource, error) {
	issuer := "https://accounts.google.com"
	secretsFile := "/Users/jlewi/secrets/bytetoko-tff-sheets-oauth.json"
	flow, err := auth.NewOIDCWebFlowHelper(secretsFile, issuer)
	if err != nil {
		return nil, err
	}

	ts, err := flow.GetTokenSource(context.Background())
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get tokensource")
	}

	// Get an initial token to make sure the flow works.
	if _, err := ts.Token(); err != nil {
		return nil, errors.Wrapf(err, "Failed to get token")
	}

	gTs := oauth.TokenSource{TokenSource: ts}
	return &gTs, nil
}
