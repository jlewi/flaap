// package main provides the entrypoint for the gRPC server
package main

import (
	"fmt"

	"github.com/jlewi/flaap/go/cmd/commands"

	"os"

	"github.com/go-logr/logr"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/spf13/cobra"
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
			log.Info("logger initialized", "level", level, "jsonLogs", jsonLog)
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := commands.RunServer(port, fileName)
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

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}
