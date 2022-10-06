// CLI is a command line interface.
package main

import (
	"fmt"
	"os"

	"github.com/jlewi/flaap/go/cmd/commands"

	"github.com/go-logr/logr"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/spf13/cobra"
)

var (
	log logr.Logger
)

const (
	defaultAPIEndpoint = "localhost:8081"
)

func newRootCmd() *cobra.Command {
	var level string
	var jsonLog bool
	rootCmd := &cobra.Command{
		Short: "flapp CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			newLogger, err := logging.InitLogger(level, !jsonLog)
			if err != nil {
				panic(err)
			}
			log = *newLogger
		},
	}

	rootCmd.PersistentFlags().StringVarP(&level, "level", "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", false, "Enable json logging.")
	return rootCmd
}

func main() {
	rootCmd := newRootCmd()
	getCmd := commands.NewGetCmd()
	getCmd.AddCommand(commands.NewGetTasksCmd())
	getCmd.AddCommand(commands.NewGetStatusCmd())
	rootCmd.AddCommand(getCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}
