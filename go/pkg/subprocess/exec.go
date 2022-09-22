package subprocess

import (
	"io"
	"strings"

	gocmd "github.com/go-cmd/cmd"
	"github.com/go-logr/logr"
	"github.com/jlewi/p22h/backend/pkg/logging"
	"github.com/pkg/errors"
)

// ExecHelper is a wrapper for executing shell commands.
//
// TODO(jeremy): Should we get rid of the methods that don't use "github.com/go-cmd/cmd" i.e RunCommands,
// Run, and RunQuietly? Should we refactor them so we use go-cmd.Cmd everywhere?
type ExecHelper struct {
	Log logr.Logger
}

// RunStreaming executes a command asynchronously and streams the stdout
// and stderr to the writer.
//
// Important: There is no guarantee that stdout and stderr will be written in the same order in which they
// are invoked due to buffering.
//
// The cmd must have streaming enabled; e.g. it can be created
// using NewCmdOptions and setting the streaming option to true.
func (h *ExecHelper) RunStreaming(cmd *gocmd.Cmd, w io.Writer) <-chan gocmd.Status {
	if cmd.Stdout == nil || cmd.Stderr == nil {
		status := gocmd.Status{
			// Set a nonzero exit code so that if code just checks exit it detects non-successful termination
			Exit:  1,
			Error: errors.New("command can't be streamed. cmd.Stdout or cmd.Stderr is nil; was the cmd created with streaming enabled? Use go-cmd.NewCmdOptions to create a command with streaming enabled."),
		}
		statusChan := make(chan gocmd.Status, 1)
		statusChan <- status
		return statusChan
	}

	// Important we need to start receiving before starting the command.
	go func() {
		for line := range cmd.Stdout {
			if _, err := w.Write([]byte(line + "\n")); err != nil {
				h.Log.Error(err, "Failed to write bytes to command's stdout", "name", cmd.Name)
			}
		}
	}()

	go func() {
		for line := range cmd.Stderr {
			if _, err := w.Write([]byte(line + "\n")); err != nil {
				h.Log.Error(err, "Failed to write bytes to commands stderr", "name", cmd.Name)
			}
		}
	}()

	h.Log.V(logging.Debug).Info("Starting command", "command", ToString(cmd), "dir", cmd.Dir, "env", cmd.Env)
	return cmd.Start()
}

func ToString(cmd *gocmd.Cmd) string {
	return cmd.Name + " " + strings.Join(cmd.Args, " ")
}
